package rust

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

const asyncClientTemplate = `{{.Header}}
use crate::error::{ProblemJson, ViiperError};
use crate::types::*;
use std::net::SocketAddr;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::TcpStream;
use tokio::net::tcp::{OwnedReadHalf, OwnedWriteHalf};

/// VIIPER management API client (asynchronous).
#[cfg(feature = "async")]
pub struct AsyncViiperClient {
    addr: SocketAddr,
}

#[cfg(feature = "async")]
impl AsyncViiperClient {
    /// Create a new async VIIPER client connecting to the specified address.
    pub fn new(addr: SocketAddr) -> Self {
        Self { addr }
    }

    async fn do_request<T: for<'de> serde::Deserialize<'de>>(
        &self,
        path: &str,
        payload: Option<&str>,
    ) -> Result<T, ViiperError> {
        let mut stream = TcpStream::connect(self.addr).await?;
        stream.set_nodelay(true)?;

        stream.write_all(path.as_bytes()).await?;
        if let Some(p) = payload {
            stream.write_all(b" ").await?;
            stream.write_all(p.as_bytes()).await?;
        }
        stream.write_all(b"\0").await?;

        let mut buf = Vec::new();
        stream.read_to_end(&mut buf).await?;

        let response = String::from_utf8(buf)
            .map_err(|_| ViiperError::UnexpectedResponse("invalid UTF-8".into()))?
            .trim_end_matches('\n')
            .to_string();

        if response.starts_with("{\"status\":") {
            let problem: ProblemJson = serde_json::from_str(&response)?;
            return Err(ViiperError::Protocol(problem));
        }

        serde_json::from_str(&response).map_err(Into::into)
    }
{{range .Routes}}{{if eq .Method "Register"}}
    /// {{.Handler}}: {{.Path}}{{if .ResponseDTO}} -> {{.ResponseDTO}}{{end}}
    pub async fn {{toSnakeCase .Handler}}(&self{{generateMethodParamsRust .}}) -> Result<{{if .ResponseDTO}}{{.ResponseDTO}}{{else}}(){{end}}, ViiperError> {
        let path = {{generatePathRust .}};
        {{generatePayloadRust .}}
        {{if .ResponseDTO}}self.do_request(&path, payload.as_deref()).await{{else}}self.do_request::<serde_json::Value>(&path, payload.as_deref()).await?;
        Ok(()){{end}}
    }
{{end}}{{end}}
    /// Connect to a device stream for sending input and receiving output.
    pub async fn connect_device(&self, bus_id: u32, dev_id: &str) -> Result<AsyncDeviceStream, ViiperError> {
        AsyncDeviceStream::connect(self.addr, bus_id, dev_id).await
    }
}

/// An async connected device stream for bidirectional communication.
/// Uses split read/write halves to allow simultaneous sending and receiving.
#[cfg(feature = "async")]
pub struct AsyncDeviceStream {
    reader: std::sync::Arc<tokio::sync::Mutex<OwnedReadHalf>>,
    writer: std::sync::Arc<tokio::sync::Mutex<OwnedWriteHalf>>,
    cancel_token: Option<tokio_util::sync::CancellationToken>,
    disconnect_callback: std::sync::Mutex<Option<Box<dyn FnOnce() + Send + 'static>>>,
}

#[cfg(feature = "async")]
impl AsyncDeviceStream {
    pub async fn connect(addr: SocketAddr, bus_id: u32, dev_id: &str) -> Result<Self, ViiperError> {
        let mut stream = TcpStream::connect(addr).await?;
		stream.set_nodelay(true)?;
		let handshake = format!("bus/{}/{}\0", bus_id, dev_id);
        stream.write_all(handshake.as_bytes()).await?;
        let (reader, writer) = stream.into_split();
        Ok(Self { 
            reader: std::sync::Arc::new(tokio::sync::Mutex::new(reader)),
            writer: std::sync::Arc::new(tokio::sync::Mutex::new(writer)),
            cancel_token: None,
            disconnect_callback: std::sync::Mutex::new(None),
        })
    }

    /// Send a device input to the device.
    pub async fn send<T: crate::wire::DeviceInput>(
        &self,
        input: &T,
    ) -> Result<(), ViiperError> {
        let bytes = input.to_bytes();
        let mut writer = self.writer.lock().await;
        writer.write_all(&bytes).await?;
        Ok(())
    }

    /// Send a device input to the device with a timeout.
    ///
    /// # Arguments
    /// * ` + "`" + `input` + "`" + ` - The device input to send
    /// * ` + "`" + `timeout` + "`" + ` - Timeout duration for the operation
    pub async fn send_timeout<T: crate::wire::DeviceInput>(
        &self,
        input: &T,
        timeout: std::time::Duration,
    ) -> Result<(), ViiperError> {
        let bytes = input.to_bytes();
        let mut writer = self.writer.lock().await;
        tokio::time::timeout(timeout, writer.write_all(&bytes))
            .await
            .map_err(|_| ViiperError::Timeout)?
            .map_err(Into::into)
    }

    /// Register a callback to receive device output asynchronously.
    /// The callback receives the read half of the stream and must read the exact number of bytes expected.
    /// The callback will be invoked repeatedly on a tokio task until it returns an error.
    /// Only one callback can be registered at a time.
    pub fn on_output<F, Fut>(&mut self, callback: F) -> Result<(), ViiperError>
    where
        F: Fn(std::sync::Arc<tokio::sync::Mutex<OwnedReadHalf>>) -> Fut + Send + 'static,
        Fut: std::future::Future<Output = std::io::Result<()>> + Send + 'static,
    {
        if self.cancel_token.is_some() {
            return Err(ViiperError::UnexpectedResponse("Output callback already registered".into()));
        }

        let reader = self.reader.clone();
        let cancel_token = tokio_util::sync::CancellationToken::new();
        let cancel_clone = cancel_token.clone();
		let Ok(mut guard) = self.disconnect_callback.lock() else {
			return Err(ViiperError::UnexpectedResponse("Disconnect callback mutex poisoned".into()));
		};
		let disconnect = guard.take();

        tokio::spawn(async move {
            loop {
                tokio::select! {
                    _ = cancel_clone.cancelled() => break,
                    result = callback(reader.clone()) => {
                        match result {
                            Ok(()) => continue,
                            Err(_) => break,
                        }
                    }
                }
            }
            if let Some(on_disconnect) = disconnect {
                on_disconnect();
            }
        });
        self.cancel_token = Some(cancel_token);
        Ok(())
    }

    pub fn on_disconnect<F>(&mut self, callback: F) -> Result<(), ViiperError>
    where
        F: FnOnce() + Send + 'static,
    {
		let Ok(mut guard) = self.disconnect_callback.lock() else {
			return Err(ViiperError::UnexpectedResponse("Disconnect callback mutex poisoned".into()));
		};
		*guard = Some(Box::new(callback));
		Ok(())
    }

    /// Send raw bytes to the device.
    pub async fn send_raw(&self, data: &[u8]) -> Result<(), ViiperError> {
        let mut writer = self.writer.lock().await;
        writer.write_all(data).await?;
        Ok(())
    }

    /// Read raw bytes from the device.
    pub async fn read_raw(&self, buf: &mut [u8]) -> Result<usize, ViiperError> {
        let mut reader = self.reader.lock().await;
        reader.read(buf).await.map_err(Into::into)
    }

    /// Read exact number of bytes from the device.
    pub async fn read_exact(&self, buf: &mut [u8]) -> Result<(), ViiperError> {
        let mut reader = self.reader.lock().await;
        reader.read_exact(buf).await?;
        Ok(())
    }
}

#[cfg(feature = "async")]
impl Drop for AsyncDeviceStream {
    fn drop(&mut self) {
        if let Some(token) = &self.cancel_token {
            token.cancel();
        }
    }
}
`

func generateAsyncClient(logger *slog.Logger, srcDir string, md *meta.Metadata) error {
	logger.Debug("Generating async_client.rs (async API)")
	outputFile := filepath.Join(srcDir, "async_client.rs")

	funcMap := template.FuncMap{
		"toSnakeCase":              common.ToSnakeCase,
		"generateMethodParamsRust": generateMethodParamsRust,
		"generatePathRust":         generatePathRust,
		"generatePayloadRust":      generatePayloadRust,
	}

	tmpl, err := template.New("asyncclient").Funcs(funcMap).Parse(asyncClientTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	data := struct {
		Header string
		Routes []scanner.RouteInfo
	}{
		Header: writeFileHeaderRust(),
		Routes: md.Routes,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	logger.Info("Generated async_client.rs", "file", outputFile)
	return nil
}
