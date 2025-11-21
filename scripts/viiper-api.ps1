#vibe coded bullshit, but works
function Invoke-ViiperApi {
    <#
    .SYNOPSIS
        Send commands to the VIIPER API server over TCP.
    
    .DESCRIPTION
        Sends a command to the VIIPER API server and returns the response.
    
    .PARAMETER Command
        The API command to send (e.g., "bus.list", "bus.create MyBus")
    
    .PARAMETER Port
        The TCP port where the API server is listening. Default: 3242
    
    .PARAMETER Hostname
        The hostname or IP address of the API server. Default: localhost
    
    .EXAMPLE
        Invoke-ViiperApi "bus.list"
    
    .EXAMPLE
        Invoke-ViiperApi "bus.create MyBus" -Port 3242
    
    .EXAMPLE
        Invoke-ViiperApi "device.add MyBus xbox360" -Hostname "192.168.1.100"
    #>
    param(
        [Parameter(Mandatory=$true, Position=0)]
        [string]$Command,
        
        [Parameter(Mandatory=$false)]
        [int]$Port = 3242,
        
        [Parameter(Mandatory=$false)]
        [string]$Hostname = "localhost"
    )
    
    try {
        $client = New-Object System.Net.Sockets.TcpClient($Hostname, $Port)
        $stream = $client.GetStream()
        $writer = New-Object System.IO.StreamWriter($stream)
        $reader = New-Object System.IO.StreamReader($stream)
        
        # Send command with double newline delimiter
        $writer.Write($Command)
        $writer.Write("`n`n")
        $writer.Flush()
        
        # Read single line response
        $response = $reader.ReadLine()
        
        $client.Close()
        
        return $response.TrimEnd("`n")
    }
    catch {
        Write-Error "Failed to connect to VIIPER API at ${Hostname}:${Port} - $_"
    }
}

function Connect-ViiperDevice {
    <#
    .SYNOPSIS
        Open a persistent USB-IP connection to a VIIPER device.
    
    .DESCRIPTION
        Connects to the VIIPER USB-IP server and keeps the connection open until Ctrl+C.
        This is useful for testing device connections and seeing when they close on device removal.
    
    .PARAMETER BusID
        The device bus ID to connect to (e.g., "1-1")
    
    .PARAMETER Port
        The USB-IP server port. Default: 3240
    
    .PARAMETER Hostname
        The hostname or IP address of the USB-IP server. Default: localhost
    
    .EXAMPLE
        Connect-ViiperDevice "1-1"
    
    .EXAMPLE
        Connect-ViiperDevice "1-1" -Port 3240 -Hostname "192.168.1.100"
    #>
    param(
        [Parameter(Mandatory=$true, Position=0)]
        [string]$BusID,
        
        [Parameter(Mandatory=$false)]
        [int]$Port = 3240,
        
        [Parameter(Mandatory=$false)]
        [string]$Hostname = "localhost"
    )
    
    try {
        Write-Host "Connecting to USB-IP device $BusID at ${Hostname}:${Port}..." -ForegroundColor Cyan
        Write-Host "Press Ctrl+C to disconnect" -ForegroundColor Yellow
        
        $client = New-Object System.Net.Sockets.TcpClient($Hostname, $Port)
        $stream = $client.GetStream()
        
        # Send OP_REQ_IMPORT (0x8003) for the specified device
        $writer = New-Object System.IO.BinaryWriter($stream)
        
        # USB-IP header: version (0x0111), command (0x8003), status (0)
        $writer.Write([uint16]0x0111)  # version (big-endian will be swapped)
        $writer.Write([uint16]0x8003)  # OP_REQ_IMPORT
        $writer.Write([uint32]0)       # status
        
        # Write busid (32 bytes, null-terminated)
        $busidBytes = [System.Text.Encoding]::ASCII.GetBytes($BusID)
        $writer.Write($busidBytes)
        # Pad with zeros to 32 bytes
        for ($i = $busidBytes.Length; $i -lt 32; $i++) {
            $writer.Write([byte]0)
        }
        
        $writer.Flush()
        
        Write-Host "Import request sent, connection established" -ForegroundColor Green
        Write-Host "Waiting for device removal or Ctrl+C..." -ForegroundColor Gray
        
        # Keep connection open and monitor for closure
        while ($client.Connected -and $stream.CanRead) {
            if ($stream.DataAvailable) {
                $buffer = New-Object byte[] 1024
                $read = $stream.Read($buffer, 0, 1024)
                if ($read -eq 0) {
                    Write-Host "`nConnection closed by server (device may have been removed)" -ForegroundColor Yellow
                    break
                }
                # Could log received data here if needed
            }
            Start-Sleep -Milliseconds 100
        }
        
        $client.Close()
        Write-Host "Disconnected" -ForegroundColor Gray
    }
    catch [System.Net.Sockets.SocketException] {
        Write-Host "`nConnection closed by server (device removed or server stopped)" -ForegroundColor Yellow
    }
    catch {
        Write-Error "Failed to connect to USB-IP server at ${Hostname}:${Port} - $_"
    }
}

# Alias for shorter command
Set-Alias -Name viiper -Value Invoke-ViiperApi

Write-Host "VIIPER API helper loaded. Usage:" -ForegroundColor Green
Write-Host "  Invoke-ViiperApi 'bus/list'" -ForegroundColor Cyan
Write-Host "  viiper 'bus/create'" -ForegroundColor Cyan
Write-Host "  viiper 'bus/1/add xbox360' -Port 3242" -ForegroundColor Cyan
Write-Host "  viiper 'bus/1/remove 1'" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Connect-ViiperDevice '1-1'          # Keep connection open until Ctrl+C" -ForegroundColor Cyan
