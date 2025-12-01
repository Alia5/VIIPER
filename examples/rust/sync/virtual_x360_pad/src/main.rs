use std::thread;
use std::time::Duration;
use std::net::ToSocketAddrs;
use viiper_client::{devices::xbox360::*, ViiperClient};

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 2 {
        eprintln!("Usage: {} <api_addr>", args[0]);
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
            r#type: Some("xbox360".to_string()),
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
            let left = buf[0];
            let right = buf[1];
            println!("← Rumble: Left={}, Right={}", left, right);
            Ok(())
        })
        .expect("Failed to register rumble callback");

    // Send controller inputs at 60fps (16ms intervals)
    let mut frame = 0u64;

    loop {
        frame += 1;

        let buttons = match (frame / 60) % 4 {
            0 => BUTTON_A,
            1 => BUTTON_B,
            2 => BUTTON_X,
            _ => BUTTON_Y,
        };

        let state = Xbox360Input {
            buttons: buttons as u32,
            lt: ((frame * 2) % 256) as u8,
            rt: ((frame * 3) % 256) as u8,
            lx: (20000.0 * 0.7071) as i16,
            ly: (20000.0 * 0.7071) as i16,
            rx: 0,
            ry: 0,
        };

        if let Err(e) = stream.send(&state) {
            eprintln!("Write error: {}", e);
            break;
        }

        if frame % 60 == 0 {
            println!(
                "→ Sent input (frame {}): buttons=0x{:04x}, LT={}, RT={}",
                frame, state.buttons, state.lt, state.rt
            );
        }

        thread::sleep(Duration::from_millis(16));
    }

    // Cleanup
    let _ = client.bus_device_remove(device_info.bus_id, Some(&device_info.dev_id));
    if created_bus {
        let _ = client.bus_remove(Some(bus_id));
    }
}
