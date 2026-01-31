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
use tokio::io::{AsyncRead, AsyncReadExt, AsyncWrite, AsyncWriteExt};
use tokio::net::TcpStream;

/// Stream wrapper that can be either plain or encrypted
#[cfg(feature = "async")]
pub enum AsyncStreamWrapper {
    Plain(TcpStream),
    Encrypted(crate::auth::AsyncEncryptedStream),
}

/// Read-half wrapper that can be either plain or encrypted
#[cfg(feature = "async")]
pub enum AsyncReadWrapper {
    Plain(tokio::net::tcp::OwnedReadHalf),
    Encrypted(crate::auth::AsyncEncryptedRead),
}

/// Write-half wrapper that can be either plain or encrypted
#[cfg(feature = "async")]
pub enum AsyncWriteWrapper {
    Plain(tokio::net::tcp::OwnedWriteHalf),
    Encrypted(crate::auth::AsyncEncryptedWrite),
}

#[cfg(feature = "async")]
impl AsyncRead for AsyncStreamWrapper {
    fn poll_read(
        mut self: std::pin::Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
        buf: &mut tokio::io::ReadBuf<'_>,
    ) -> std::task::Poll<std::io::Result<()>> {
        match &mut *self {
            AsyncStreamWrapper::Plain(s) => std::pin::Pin::new(s).poll_read(cx, buf),
            AsyncStreamWrapper::Encrypted(s) => std::pin::Pin::new(s).poll_read(cx, buf),
        }
    }
}

#[cfg(feature = "async")]
impl AsyncRead for AsyncReadWrapper {
    fn poll_read(
        mut self: std::pin::Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
        buf: &mut tokio::io::ReadBuf<'_>,
    ) -> std::task::Poll<std::io::Result<()>> {
        match &mut *self {
            AsyncReadWrapper::Plain(s) => std::pin::Pin::new(s).poll_read(cx, buf),
            AsyncReadWrapper::Encrypted(s) => std::pin::Pin::new(s).poll_read(cx, buf),
        }
    }
}

#[cfg(feature = "async")]
impl AsyncWrite for AsyncStreamWrapper {
    fn poll_write(
        mut self: std::pin::Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
        buf: &[u8],
    ) -> std::task::Poll<Result<usize, std::io::Error>> {
        match &mut *self {
            AsyncStreamWrapper::Plain(s) => std::pin::Pin::new(s).poll_write(cx, buf),
            AsyncStreamWrapper::Encrypted(s) => std::pin::Pin::new(s).poll_write(cx, buf),
        }
    }
    
    fn poll_flush(
        mut self: std::pin::Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
    ) -> std::task::Poll<Result<(), std::io::Error>> {
        match &mut *self {
            AsyncStreamWrapper::Plain(s) => std::pin::Pin::new(s).poll_flush(cx),
            AsyncStreamWrapper::Encrypted(s) => std::pin::Pin::new(s).poll_flush(cx),
        }
    }
    
    fn poll_shutdown(
        mut self: std::pin::Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
    ) -> std::task::Poll<Result<(), std::io::Error>> {
        match &mut *self {
            AsyncStreamWrapper::Plain(s) => std::pin::Pin::new(s).poll_shutdown(cx),
            AsyncStreamWrapper::Encrypted(s) => std::pin::Pin::new(s).poll_shutdown(cx),
        }
    }
}

#[cfg(feature = "async")]
impl AsyncWrite for AsyncWriteWrapper {
    fn poll_write(
        mut self: std::pin::Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
        buf: &[u8],
    ) -> std::task::Poll<Result<usize, std::io::Error>> {
        match &mut *self {
            AsyncWriteWrapper::Plain(s) => std::pin::Pin::new(s).poll_write(cx, buf),
            AsyncWriteWrapper::Encrypted(s) => std::pin::Pin::new(s).poll_write(cx, buf),
        }
    }
    
    fn poll_flush(
        mut self: std::pin::Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
    ) -> std::task::Poll<Result<(), std::io::Error>> {
        match &mut *self {
            AsyncWriteWrapper::Plain(s) => std::pin::Pin::new(s).poll_flush(cx),
            AsyncWriteWrapper::Encrypted(s) => std::pin::Pin::new(s).poll_flush(cx),
        }
    }
    
    fn poll_shutdown(
        mut self: std::pin::Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
    ) -> std::task::Poll<Result<(), std::io::Error>> {
        match &mut *self {
            AsyncWriteWrapper::Plain(s) => std::pin::Pin::new(s).poll_shutdown(cx),
            AsyncWriteWrapper::Encrypted(s) => std::pin::Pin::new(s).poll_shutdown(cx),
        }
    }
}

/// VIIPER management API client (asynchronous).
#[cfg(feature = "async")]
pub struct AsyncViiperClient {
    addr: SocketAddr,
    password: Option<String>,
}

#[cfg(feature = "async")]
impl AsyncViiperClient {
    /// Create a new async VIIPER client connecting to the specified address.
    pub fn new(addr: SocketAddr) -> Self {
        Self { addr, password: None }
    }

    /// Create a new async VIIPER client with password authentication.
    /// Empty password string explicitly means no authentication.
    pub fn new_with_password(addr: SocketAddr, password: String) -> Self {
        let password = if password.is_empty() { None } else { Some(password) };
        Self { addr, password }
    }

    async fn do_request<T: for<'de> serde::Deserialize<'de>>(
        &self,
        path: &str,
        payload: Option<&str>,
    ) -> Result<T, ViiperError> {
        let tcp_stream = TcpStream::connect(self.addr).await?;
        tcp_stream.set_nodelay(true)?;

        let mut stream = if let Some(ref pwd) = self.password {
            AsyncStreamWrapper::Encrypted(crate::auth::perform_handshake_async(tcp_stream, pwd).await?)
        } else {
            AsyncStreamWrapper::Plain(tcp_stream)
        };

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
        AsyncDeviceStream::connect(self.addr, bus_id, dev_id, self.password.as_deref()).await
    }
}

/// An async connected device stream for bidirectional communication.
#[cfg(feature = "async")]
pub struct AsyncDeviceStream {
    read_stream: std::sync::Arc<tokio::sync::Mutex<AsyncReadWrapper>>,
    write_stream: std::sync::Arc<tokio::sync::Mutex<AsyncWriteWrapper>>,
    cancel_token: Option<tokio_util::sync::CancellationToken>,
    disconnect_callback: std::sync::Mutex<Option<Box<dyn FnOnce() + Send + 'static>>>,
}

#[cfg(feature = "async")]
impl AsyncDeviceStream {
    pub async fn connect(addr: SocketAddr, bus_id: u32, dev_id: &str, password: Option<&str>) -> Result<Self, ViiperError> {
        let tcp_stream = TcpStream::connect(addr).await?;
		tcp_stream.set_nodelay(true)?;
		
        let (read_stream, mut write_stream) = if let Some(pwd) = password {
            let encrypted = crate::auth::perform_handshake_async(tcp_stream, pwd).await?;
            let (read_half, write_half) = encrypted.into_split();
            (AsyncReadWrapper::Encrypted(read_half), AsyncWriteWrapper::Encrypted(write_half))
        } else {
            let (read_half, write_half) = tcp_stream.into_split();
            (AsyncReadWrapper::Plain(read_half), AsyncWriteWrapper::Plain(write_half))
        };
        
        let handshake = format!("bus/{}/{}\0", bus_id, dev_id);
        write_stream.write_all(handshake.as_bytes()).await?;
        
        Ok(Self { 
            read_stream: std::sync::Arc::new(tokio::sync::Mutex::new(read_stream)),
            write_stream: std::sync::Arc::new(tokio::sync::Mutex::new(write_stream)),
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
        let mut stream = self.write_stream.lock().await;
        stream.write_all(&bytes).await?;
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
        let mut stream = self.write_stream.lock().await;
        tokio::time::timeout(timeout, stream.write_all(&bytes))
            .await
            .map_err(|_| ViiperError::Timeout)?
            .map_err(Into::into)
    }

    /// Register a callback to receive device output asynchronously.
    /// The callback receives a shared reference to the read half and must read the exact number of bytes expected.
    /// The callback will be invoked repeatedly on a tokio task until it returns an error.
    /// Only one callback can be registered at a time.
    pub fn on_output<F, Fut>(&mut self, callback: F) -> Result<(), ViiperError>
    where
        F: Fn(std::sync::Arc<tokio::sync::Mutex<AsyncReadWrapper>>) -> Fut + Send + 'static,
        Fut: std::future::Future<Output = std::io::Result<()>> + Send + 'static,
    {
        if self.cancel_token.is_some() {
            return Err(ViiperError::UnexpectedResponse("Output callback already registered".into()));
        }

        let stream = self.read_stream.clone();
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
                    result = callback(stream.clone()) => {
                        match result {
                            Ok(()) => continue,
                            Err(_) => break,
                        }
                    }
                }
            }
            if let Some(cb) = disconnect {
                cb();
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
        let mut stream = self.write_stream.lock().await;
        stream.write_all(data).await?;
        Ok(())
    }

    /// Read raw bytes from the device.
    pub async fn read_raw(&self, buf: &mut [u8]) -> Result<usize, ViiperError> {
        let mut stream = self.read_stream.lock().await;
        stream.read(buf).await.map_err(Into::into)
    }

    /// Read exact number of bytes from the device.
    pub async fn read_exact(&self, buf: &mut [u8]) -> Result<(), ViiperError> {
        let mut stream = self.read_stream.lock().await;
        stream.read_exact(buf).await?;
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
