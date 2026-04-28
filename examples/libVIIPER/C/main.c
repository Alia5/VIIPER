

#include <stdio.h>
#include <stdbool.h>
#include <signal.h>
#include <stdint.h>
#include <math.h>
#include "libVIIPER.h"

static volatile bool g_running = true;
static void signalHandler(int sig) { (void)sig; g_running = false; }

void rumbleCallback(Xbox360DeviceHandle handle, uint8_t leftMotor, uint8_t rightMotor) {
    (void)handle;
    printf("<- Rumble: Left=%u, Right=%u\n", leftMotor, rightMotor);
}


void logCallback(VIIPERLogLevel level, const char* message) {
    const char* levelStr;
    switch (level) {
        case VIIPER_LOG_DEBUG:
            levelStr = "DEBUG";
            break;
        case VIIPER_LOG_INFO:
            levelStr = "INFO";
            break;
        case VIIPER_LOG_WARN:
            levelStr = "WARN";
            break;
        case VIIPER_LOG_ERROR:
            levelStr = "ERROR";
            break;
        default:
            levelStr = "UNKNOWN";
    }
    printf("libVIIPER [%s] %s\n", levelStr, message);
}

int main() {
    printf("Hello, libVIIPER!\n");

    // Create a usb-server config.
    // All fields are optional, default listen address is "0.0.0.0:3241"
    USBServerConfig conf = {
        .addr = "localhost:3245",
    };
    USBServerHandle serverHandle = 0;
    // Create a new USB-Server on the specified listening address
    // The server will run in a background thread, and can be stopped by calling CloseUSBServer with the returned handle.
    // LogCallback can be NULL
    bool success = NewUSBServer(&conf, &serverHandle, logCallback);
    if (!success) {
        printf("Failed to create USB server.\n");
        return 1;
    }
    printf("Created USB server with handle: %lu\n", (unsigned long)serverHandle);


    uint32_t busID = 0;
    // Create a new USB bus with the specified bus ID, or automatically assign one if busID is 0.
    success = CreateUSBBus(serverHandle, &busID);
    if (!success) {
        printf("Failed to create USB bus.\n");
        return 1;
    }
    printf("Created USB bus with ID: %u\n", busID);


    Xbox360DeviceHandle deviceHandle = 0;
    success = CreateXbox360Device(serverHandle, &deviceHandle, busID, true, 0,0,0);
    if (!success) {
        printf("Failed to create Xbox 360 device.\n");
        return 1;
    }

    signal(SIGINT,  signalHandler);
    signal(SIGTERM, signalHandler);

    SetXbox360RumbleCallback(deviceHandle, rumbleCallback);

    _sleep(5000);

    Xbox360DeviceState deviceState = {0};
    uint64_t frame = 0;

    while (g_running) {
        frame++;
        switch ((frame / 60) % 4) {
            case 0:
                deviceState.Buttons = XBOX360_BUTTON_A;
                break;
            case 1: 
                deviceState.Buttons = XBOX360_BUTTON_B;
                break;
            case 2:
                deviceState.Buttons = XBOX360_BUTTON_X;
                break;
            default:
                deviceState.Buttons = XBOX360_BUTTON_Y;
                break;
        }
        deviceState.LT = (uint8_t)((frame * 2) % 256);
        deviceState.RT = (uint8_t)((frame * 3) % 256);
        deviceState.LX = (int16_t)(20000.0 * sin(frame * 0.02));
        deviceState.LY = (int16_t)(20000.0 * cos(frame * 0.02));

        SetXbox360DeviceState(deviceHandle, deviceState);

        if (frame % 60 == 0) {
            printf("-> Sent input (frame %llu): buttons=0x%04X, LT=%u, RT=%u\n",
                (unsigned long long)frame, deviceState.Buttons, deviceState.LT, deviceState.RT);
        }

        _sleep(16);
    }

    printf("\nShutting down...\n");
    CloseUSBServer(serverHandle);
    return 0;
}