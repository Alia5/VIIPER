use std::net::ToSocketAddrs;
use std::thread;
use std::time::Duration;
use viiper_client::{devices::keyboard::*, ViiperClient};

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 2 {
        eprintln!("Usage: {} <host>", args[0]);
        eprintln!("Example: {} localhost:3242", args[0]);
        std::process::exit(1);
    }

    let addr_str = &args[1];
    let addr: std::net::SocketAddr = match addr_str.to_socket_addrs() {
        Ok(mut iter) => match iter.next() {
            Some(a) => a,
            None => {
                eprintln!("Invalid address '{}': no resolvable addresses", addr_str);
                std::process::exit(1);
            }
        },
        Err(e) => {
            eprintln!("Invalid address '{}': {}", addr_str, e);
            std::process::exit(1);
        }
    };

    let client = ViiperClient::new(addr);

    // Find or create a bus
    let (bus_id, created_bus) = match client.bus_list() {
        Ok(resp) if resp.buses.is_empty() => match client.bus_create(None) {
            Ok(r) => {
                println!("Created bus {}", r.bus_id);
                (r.bus_id, true)
            }
            Err(e) => {
                eprintln!("BusCreate failed: {}", e);
                std::process::exit(1);
            }
        },
        Ok(resp) => {
            let bus_id = *resp.buses.iter().min().unwrap();
            println!("Using existing bus {}", bus_id);
            (bus_id, false)
        }
        Err(e) => {
            eprintln!("BusList error: {}", e);
            std::process::exit(1);
        }
    };

    // Add device
    let device_info = match client.bus_device_add(
        bus_id,
        &viiper_client::types::DeviceCreateRequest {
            r#type: Some("keyboard".to_string()),
            id_vendor: None,
            id_product: None,
        },
    ) {
        Ok(d) => d,
        Err(e) => {
            eprintln!("AddDevice error: {}", e);
            if created_bus {
                let _ = client.bus_remove(Some(bus_id));
            }
            std::process::exit(1);
        }
    };

    // Connect to device stream
    let mut stream = match client.connect_device(device_info.bus_id, &device_info.dev_id) {
        Ok(s) => s,
        Err(e) => {
            eprintln!("ConnectDevice error: {}", e);
            let _ = client.bus_device_remove(device_info.bus_id, Some(&device_info.dev_id));
            if created_bus {
                let _ = client.bus_remove(Some(bus_id));
            }
            std::process::exit(1);
        }
    };

    println!(
        "Created and connected to device {} on bus {}",
        device_info.dev_id, device_info.bus_id
    );

    stream
        .on_disconnect(|| {
            eprintln!("Device disconnected by server");
            std::process::exit(0);
        })
        .expect("Failed to register disconnect callback");

    stream
        .on_output(|reader| {
            let mut buf = [0u8; OUTPUT_SIZE];
            reader.read_exact(&mut buf)?;
            let leds = buf[0];
            let num_lock = (leds & 0x01) != 0;
            let caps_lock = (leds & 0x02) != 0;
            let scroll_lock = (leds & 0x04) != 0;
            let compose = (leds & 0x08) != 0;
            let kana = (leds & 0x10) != 0;
            println!(
                "← LEDs: Num={} Caps={} Scroll={} Compose={} Kana={}",
                num_lock, caps_lock, scroll_lock, compose, kana
            );
            Ok(())
        })
        .expect("Failed to register LED callback");

    println!("Every 5s: type 'Hello!' + Enter. Press Ctrl+C to stop.");

    // Type "Hello!" + Enter every 5 seconds
    loop {
        if let Err(e) = type_string(&mut stream, "Hello!") {
            eprintln!("Write error: {}", e);
            break;
        }

        thread::sleep(Duration::from_millis(100));

        if let Err(e) = press_key(&mut stream, KEY_ENTER) {
            eprintln!("Write error: {}", e);
            break;
        }

        println!("→ Typed: Hello!");
        thread::sleep(Duration::from_secs(5));
    }

    // Cleanup
    let _ = client.bus_device_remove(device_info.bus_id, Some(&device_info.dev_id));
    if created_bus {
        let _ = client.bus_remove(Some(bus_id));
    }
}

fn type_string(
    stream: &mut viiper_client::DeviceStream,
    text: &str,
) -> Result<(), viiper_client::error::ViiperError> {
    for ch in text.chars() {
        let code_point = ch as u32;
        let key = match CHAR_TO_KEY.get(&(code_point as u8)) {
            Some(&k) => k,
            None => continue,
        };

        let mut mods = 0;
        if SHIFT_CHARS.contains(&(code_point as u8)) {
            mods = MOD_LEFT_SHIFT;
        }

        // Key down
        let down = KeyboardInput {
            modifiers: mods,
            count: 1,
            keys: vec![key],
        };
        stream.send(&down)?;
        thread::sleep(Duration::from_millis(100));

        // Key up
        let up = KeyboardInput {
            modifiers: 0,
            count: 0,
            keys: vec![],
        };
        stream.send(&up)?;
        thread::sleep(Duration::from_millis(100));
    }
    Ok(())
}

fn press_key(
    stream: &mut viiper_client::DeviceStream,
    key: u8,
) -> Result<(), viiper_client::error::ViiperError> {
    let press = KeyboardInput {
        modifiers: 0,
        count: 1,
        keys: vec![key],
    };
    stream.send(&press)?;
    thread::sleep(Duration::from_millis(100));

    let release = KeyboardInput {
        modifiers: 0,
        count: 0,
        keys: vec![],
    };
    stream.send(&release)?;
    Ok(())
}
