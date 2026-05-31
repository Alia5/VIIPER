/*
 * NS2Pro CLI example using libVIIPER.
 *
 * Creates a USB server + NS2Pro device and lets you set its state
 * interactively via stdin commands.
 *
 * Usage:
 *   ns2pro_cli
 *
 * Commands (case-insensitive):
 *   A=true / B=false / X=1 / Y=0 ...
 *   L=true, ZL=true, R=true, ZR=true
 *   Plus=true, Minus=true, Home=true, Capture=true
 *   Up=true, Down=true, Left=true, Right=true   (D-Pad)
 *   LS=true, RS=true                             (stick clicks)
 *   GL=true, GR=true, C=true, Headset=true
 *   LX=2048   LY=2048   RX=2048   RY=2048       (0..4095, center=2048)
 *   AccelX=0  AccelY=0  AccelZ=0
 *   GyroX=0   GyroY=0   GyroZ=0
 *   print | reset | help | quit
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <stdint.h>
#include <stdbool.h>
#include <ctype.h>

#include "libVIIPER.h"

#ifdef _WIN32
#include <windows.h>
#define strcasecmp _stricmp
#else
#include <unistd.h>
#endif

/* ── signal ──────────────────────────────────────────────────────────────── */

static volatile bool g_running = true;
static void on_signal(int sig)
{
    (void)sig;
    g_running = false;
}

/* ── callbacks ───────────────────────────────────────────────────────────── */

static void output_callback(NS2ProDeviceHandle handle, NS2ProOutputState out)
{
    (void)handle;
    printf("<- Output: PlayerLED=0x%02X Flags=0x%02X\n",
           out.PlayerLedMask, out.Flags);
}

static void log_callback(VIIPERLogLevel level, const char *msg)
{
    const char *lv =
        level == VIIPER_LOG_DEBUG ? "DEBUG" : level == VIIPER_LOG_INFO ? "INFO"
                                          : level == VIIPER_LOG_WARN   ? "WARN"
                                          : level == VIIPER_LOG_ERROR  ? "ERROR"
                                                                       : "?";
    printf("[libVIIPER/%s] %s\n", lv, msg);
}

/* ── helpers ─────────────────────────────────────────────────────────────── */

static void str_tolower(char *dst, const char *src, size_t n)
{
    size_t i;
    for (i = 0; i < n - 1 && src[i]; i++)
        dst[i] = (char)tolower((unsigned char)src[i]);
    dst[i] = '\0';
}

/* returns 1=true, 0=false, -1=parse error */
static int parse_bool(const char *v)
{
    if (!v)
        return -1;
    if (strcmp(v, "1") == 0 || strcasecmp(v, "true") == 0 ||
        strcasecmp(v, "yes") == 0 || strcasecmp(v, "on") == 0)
        return 1;
    if (strcmp(v, "0") == 0 || strcasecmp(v, "false") == 0 ||
        strcasecmp(v, "no") == 0 || strcasecmp(v, "off") == 0)
        return 0;
    return -1;
}

static uint16_t clamp_stick(long v)
{
    if (v < (long)NS2PRO_STICK_MIN)
        return NS2PRO_STICK_MIN;
    if (v > (long)NS2PRO_STICK_MAX)
        return NS2PRO_STICK_MAX;
    return (uint16_t)v;
}

static void set_button(NS2ProDeviceState *s, uint32_t mask, bool on)
{
    if (on)
        s->Buttons |= mask;
    else
        s->Buttons &= ~mask;
}

/* ── command table ───────────────────────────────────────────────────────── */

static const struct
{
    const char *name;
    uint32_t mask;
} k_buttons[] = {
    {"a", NS2PRO_BUTTON_A},
    {"b", NS2PRO_BUTTON_B},
    {"x", NS2PRO_BUTTON_X},
    {"y", NS2PRO_BUTTON_Y},
    {"l", NS2PRO_BUTTON_L},
    {"zl", NS2PRO_BUTTON_ZL},
    {"r", NS2PRO_BUTTON_R},
    {"zr", NS2PRO_BUTTON_ZR},
    {"plus", NS2PRO_BUTTON_PLUS},
    {"minus", NS2PRO_BUTTON_MINUS},
    {"home", NS2PRO_BUTTON_HOME},
    {"capture", NS2PRO_BUTTON_CAPTURE},
    {"up", NS2PRO_BUTTON_UP},
    {"down", NS2PRO_BUTTON_DOWN},
    {"left", NS2PRO_BUTTON_LEFT},
    {"right", NS2PRO_BUTTON_RIGHT},
    {"ls", NS2PRO_BUTTON_LEFT_STICK},
    {"leftstick", NS2PRO_BUTTON_LEFT_STICK},
    {"rs", NS2PRO_BUTTON_RIGHT_STICK},
    {"rightstick", NS2PRO_BUTTON_RIGHT_STICK},
    {"gl", NS2PRO_BUTTON_GL},
    {"gr", NS2PRO_BUTTON_GR},
    {"c", NS2PRO_BUTTON_C},
    {"headset", NS2PRO_BUTTON_HEADSET},
};

/* returns 0 on success, 1 on error */
static int apply_command(NS2ProDeviceState *s, const char *raw_key, const char *val)
{
    char k[64];
    str_tolower(k, raw_key, sizeof(k));

    /* buttons */
    for (size_t i = 0; i < sizeof(k_buttons) / sizeof(k_buttons[0]); i++)
    {
        if (strcmp(k, k_buttons[i].name) == 0)
        {
            int b = parse_bool(val);
            if (b < 0)
            {
                printf("Expected bool for '%s', got '%s'\n", raw_key, val);
                return 1;
            }
            set_button(s, k_buttons[i].mask, (bool)b);
            return 0;
        }
    }

    /* sticks and IMU */
    char *end;
    long lv = strtol(val, &end, 10);
    if (*end != '\0')
    {
        printf("Expected number for '%s', got '%s'\n", raw_key, val);
        return 1;
    }

    if (strcmp(k, "lx") == 0)
    {
        s->LX = clamp_stick(lv);
        return 0;
    }
    if (strcmp(k, "ly") == 0)
    {
        s->LY = clamp_stick(lv);
        return 0;
    }
    if (strcmp(k, "rx") == 0)
    {
        s->RX = clamp_stick(lv);
        return 0;
    }
    if (strcmp(k, "ry") == 0)
    {
        s->RY = clamp_stick(lv);
        return 0;
    }
    if (strcmp(k, "accelx") == 0)
    {
        s->AccelX = (int16_t)lv;
        return 0;
    }
    if (strcmp(k, "accely") == 0)
    {
        s->AccelY = (int16_t)lv;
        return 0;
    }
    if (strcmp(k, "accelz") == 0)
    {
        s->AccelZ = (int16_t)lv;
        return 0;
    }
    if (strcmp(k, "gyrox") == 0)
    {
        s->GyroX = (int16_t)lv;
        return 0;
    }
    if (strcmp(k, "gyroy") == 0)
    {
        s->GyroY = (int16_t)lv;
        return 0;
    }
    if (strcmp(k, "gyroz") == 0)
    {
        s->GyroZ = (int16_t)lv;
        return 0;
    }

    printf("Unknown key: '%s' (try 'help')\n", raw_key);
    return 1;
}

static void print_help(void)
{
    printf("Buttons (bool: true/false/1/0/yes/no/on/off):\n");
    printf("  A, B, X, Y, L, ZL, R, ZR\n");
    printf("  Plus, Minus, Home, Capture\n");
    printf("  Up, Down, Left, Right\n");
    printf("  LS / LeftStick,  RS / RightStick\n");
    printf("  GL, GR, C, Headset\n");
    printf("Sticks (0..%u, center=%u):\n", NS2PRO_STICK_MAX, NS2PRO_STICK_CENTER);
    printf("  LX, LY, RX, RY\n");
    printf("IMU (int16):\n");
    printf("  AccelX, AccelY, AccelZ\n");
    printf("  GyroX,  GyroY,  GyroZ\n");
    printf("Other:\n");
    printf("  print   show current state\n");
    printf("  reset   zero everything, re-center sticks\n");
    printf("  help    this message\n");
    printf("  quit    exit\n");
}

static void print_state(const NS2ProDeviceState *s)
{
    printf("Buttons : 0x%06X\n", s->Buttons);
    printf("Sticks  : LX=%-5u LY=%-5u RX=%-5u RY=%-5u\n",
           s->LX, s->LY, s->RX, s->RY);
    printf("Accel   : X=%-6d Y=%-6d Z=%-6d\n", s->AccelX, s->AccelY, s->AccelZ);
    printf("Gyro    : X=%-6d Y=%-6d Z=%-6d\n", s->GyroX, s->GyroY, s->GyroZ);
}

static NS2ProDeviceState default_state(void)
{
    NS2ProDeviceState s = {0};
    s.LX = s.LY = s.RX = s.RY = NS2PRO_STICK_CENTER;
    return s;
}

/* ── main ────────────────────────────────────────────────────────────────── */

int main(void)
{
    signal(SIGINT, on_signal);
    signal(SIGTERM, on_signal);

    /* create server */
    USBServerConfig conf = {.addr = "localhost:3249"};
    USBServerHandle server = 0;
    if (!NewUSBServer(&conf, &server, log_callback))
    {
        fprintf(stderr, "Failed to create USB server\n");
        return 1;
    }
    printf("USB server started on localhost:3249\n");

    /* create bus */
    uint32_t busID = 0;
    if (!CreateUSBBus(server, &busID))
    {
        fprintf(stderr, "Failed to create USB bus\n");
        CloseUSBServer(server);
        return 1;
    }
    printf("Created bus %u\n", busID);

    /* create NS2Pro device, auto-attach to local USBIP driver */
    NS2ProDeviceHandle device = 0;
    NS2ProMetaState meta = {
        .SerialNumber = "OVERRIDE-SN-00"
    };
    if (!CreateNS2ProDevice(
            server,
            &device,
            busID, /*autoAttach=*/true,
            0, 0, &meta))
    {
        fprintf(stderr, "Failed to create NS2Pro device\n");
        CloseUSBServer(server);
        return 1;
    }
    printf("NS2Pro device created\n\n");

    SetNS2ProOutputCallback(device, output_callback);

    printf("NS2Pro CLI ready. Type 'help' for commands, Ctrl+C or 'quit' to exit.\n");

    NS2ProDeviceState state = default_state();
    SetNS2ProDeviceState(device, state);

    char line[256];
    while (g_running)
    {
        printf("> ");
        fflush(stdout);

        if (!fgets(line, sizeof(line), stdin) || !g_running)
            break;

        line[strcspn(line, "\r\n")] = '\0';
        if (line[0] == '\0')
            continue;

        char low[256];
        str_tolower(low, line, sizeof(low));

        if (strcmp(low, "quit") == 0 || strcmp(low, "exit") == 0)
            break;
        if (strcmp(low, "help") == 0 || strcmp(low, "?") == 0)
        {
            print_help();
            continue;
        }
        if (strcmp(low, "print") == 0)
        {
            print_state(&state);
            continue;
        }
        if (strcmp(low, "reset") == 0)
        {
            state = default_state();
            SetNS2ProDeviceState(device, state);
            printf("State reset\n");
            continue;
        }

        char *eq = strchr(line, '=');
        if (!eq)
        {
            printf("Unknown command '%s' (try 'help')\n", line);
            continue;
        }
        *eq = '\0';
        const char *key = line;
        const char *val = eq + 1;

        if (apply_command(&state, key, val) == 0)
            SetNS2ProDeviceState(device, state);
    }

    printf("\nShutting down...\n");
    RemoveNS2ProDevice(device);
    CloseUSBServer(server);
    return 0;
}
