use std::net::ToSocketAddrs;
use std::thread;
use std::time::Duration;
use viiper_client::{devices::mouse::*, ViiperClient};

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
            r#type: Some("mouse".to_string()),
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

    println!(
        "Every 3s: move diagonally by 50px (X and Y), then click and scroll. Press Ctrl+C to stop."
    );

    // Send a short movement once every 3 seconds for easy local testing.
    // Followed by a short click and a single scroll notch.
    let mut dir = 1;
    let step = 50; // move diagonally by 50 px in X and Y (now supports up to ±32767)

    loop {
        // Move diagonally: (+step,+step) then (-step,-step) next tick
        let dx = step * dir;
        let dy = step * dir;
        dir *= -1;

        // One-shot movement report (diagonal)
        if let Err(e) = stream.send(&MouseInput {
            buttons: 0,
            dx,
            dy,
            wheel: 0,
            pan: 0,
        }) {
            eprintln!("Write error: {}", e);
            break;
        }
        println!("→ Moved mouse dx={} dy={}", dx, dy);

        // Zero state shortly after to keep movement one-shot (harmless safety)
        thread::sleep(Duration::from_millis(30));
        let _ = stream.send(&MouseInput {
            buttons: 0,
            dx: 0,
            dy: 0,
            wheel: 0,
            pan: 0,
        });

        // Simulate a short left click: press then release
        thread::sleep(Duration::from_millis(50));
        let _ = stream.send(&MouseInput {
            buttons: BTN__LEFT,
            dx: 0,
            dy: 0,
            wheel: 0,
            pan: 0,
        });
        thread::sleep(Duration::from_millis(60));
        let _ = stream.send(&MouseInput {
            buttons: 0x00,
            dx: 0,
            dy: 0,
            wheel: 0,
            pan: 0,
        });
        println!("→ Clicked (left)");

        // Simulate a short scroll: one notch upwards
        thread::sleep(Duration::from_millis(50));
        let _ = stream.send(&MouseInput {
            buttons: 0,
            dx: 0,
            dy: 0,
            wheel: 1,
            pan: 0,
        });
        thread::sleep(Duration::from_millis(30));
        let _ = stream.send(&MouseInput {
            buttons: 0,
            dx: 0,
            dy: 0,
            wheel: 0,
            pan: 0,
        });
        println!("→ Scrolled (wheel=+1)");

        thread::sleep(Duration::from_secs(3));
    }

    // Cleanup
    let _ = client.bus_device_remove(device_info.bus_id, Some(&device_info.dev_id));
    if created_bus {
        let _ = client.bus_remove(Some(bus_id));
    }
}
