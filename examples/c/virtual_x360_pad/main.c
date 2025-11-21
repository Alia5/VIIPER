#include "viiper/viiper.h"
#include "viiper/viiper_xbox360.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <time.h>

#if defined(_WIN32) || defined(_WIN64)
#  include <windows.h>
static void sleep_ms(int ms)
{
    Sleep(ms);
}
#else
#  include <unistd.h>
static void sleep_ms(int ms)
{
    usleep(ms * 1000);
}
#endif

static uint32_t choose_or_create_bus(viiper_client_t* client)
{
    viiper_bus_list_response_t list;
    memset(&list, 0, sizeof list);
    if (viiper_bus_list(client, &list) != VIIPER_OK){
        fprintf(stderr, "BusList error: %s\n", viiper_get_error(client));
        return 0;
    }
    uint32_t busId = 0;
    if (list.buses_count > 0){
        busId = list.Buses[0];
        for (size_t i = 1; i < list.buses_count; i++)
        {
            if (list.Buses[i] < busId)
            {
                busId = list.Buses[i];
            }
        }
        printf("Using existing bus %u\n", (unsigned)busId);
        viiper_free_bus_list_response(&list);
        return busId;
    }
    viiper_free_bus_list_response(&list);

    viiper_bus_create_response_t cr;
    memset(&cr, 0, sizeof cr);
    if (viiper_bus_create(client, NULL, &cr) == VIIPER_OK){
        printf("Created bus %u\n", (unsigned)cr.BusID);
        return cr.BusID;
    }
    fprintf(stderr, "BusCreate failed: %s\n", viiper_get_error(client));
    return 0;
}

static void on_rumble(const void* output, size_t output_size, void* user)
{
    (void)user;
    if (output_size >= 2) {
        const viiper_xbox360_output_t* out = (const viiper_xbox360_output_t*)output;
        printf("← Rumble: Left=%u, Right=%u\n", out->left, out->right);
    }
}

int main(int argc, char** argv)
{
    /* Args: 0 -> default 127.0.0.1:3242; 1 -> host:port; else usage */
    char host[256];
    uint16_t port = 3242;
    if (argc <= 1)
    {
        snprintf(host, sizeof host, "%s", "127.0.0.1");
    }
    else if (argc == 2)
    {
        const char* addr = argv[1];
        const char* colon = strrchr(addr, ':');
        if (!colon)
        {
            fprintf(stderr, "Usage: virtual_x360_pad [host:port]\n");
            return 1;
        }
        size_t len = (size_t)(colon - addr);
        if (len >= sizeof(host)) len = sizeof(host) - 1;
        memcpy(host, addr, len);
        host[len] = '\0';
        port = (uint16_t)atoi(colon + 1);
    }
    else
    {
        fprintf(stderr, "Usage: virtual_x360_pad [host:port]\n");
        return 1;
    }

    viiper_client_t* client = NULL;
    if (viiper_client_create(host, port, &client) != VIIPER_OK){
        fprintf(stderr, "client_create failed\n");
        return 1;
    }

    uint32_t busId = choose_or_create_bus(client);
    if (busId == 0)
    {
        viiper_client_free(client);
        return 1;
    }

    viiper_device_info_t addResp; memset(&addResp, 0, sizeof addResp);
    viiper_device_t* dev = NULL;
    const char* deviceType = "xbox360";
    viiper_device_create_request_t createReq = { .Type = &deviceType, .IdVendor = NULL, .IdProduct = NULL };
    viiper_error_t rc = viiper_add_device_and_connect(client, busId, &createReq, &addResp, &dev);
    if (rc != VIIPER_OK){
        fprintf(stderr, "add_device_and_connect failed: %s\n", viiper_get_error(client));
        viiper_client_free(client);
        return 1;
    }
    const char* devId = addResp.DevId ? addResp.DevId : "(none)";
    printf("Created and connected device %s on bus %u (type: %s)\n", devId, (unsigned)busId, addResp.Type ? addResp.Type : "unknown");

    /* Register async backchannel callback */
    viiper_device_on_output(dev, on_rumble, NULL);
    printf("Connected to device stream\n");

    unsigned long long frame = 0;
    for (;;){
        ++frame;
        viiper_xbox360_input_t in;
        memset(&in, 0, sizeof in);
        switch ((frame / 60) % 4){
            case 0:
                in.buttons = VIIPER_XBOX360_BUTTONA;
                break;
            case 1:
                in.buttons = VIIPER_XBOX360_BUTTONB;
                break;
            case 2:
                in.buttons = VIIPER_XBOX360_BUTTONX;
                break;
            default:
                in.buttons = VIIPER_XBOX360_BUTTONY;
                break;
        }
        in.lt = (uint8_t)((frame * 2) % 256);
        in.rt = (uint8_t)((frame * 3) % 256);
        in.lx = (int16_t)(20000.0 * 0.7071);
        in.ly = (int16_t)(20000.0 * 0.7071);
        in.rx = 0;
        in.ry = 0;
        rc = viiper_device_send(dev, &in, sizeof(in));
        if (rc != VIIPER_OK){
            fprintf(stderr, "send error: %s\n", viiper_get_error(client));
            break;
        }
        if (frame % 60 == 0){
            printf("→ Sent input (frame %llu): buttons=0x%04x, LT=%u, RT=%u\n", frame, (unsigned)in.buttons, in.lt, in.rt);
        }

        sleep_ms(16);
    }

    viiper_device_close(dev);
    // Optional: remove bus (will fail if devices are still present)
    viiper_bus_remove_response_t br;
    memset(&br, 0, sizeof br);
    if (viiper_bus_remove(client, busId, &br) != VIIPER_OK){
        fprintf(stderr, "BusRemove failed (continuing): %s\n", viiper_get_error(client));
    }
    viiper_client_free(client);
    return 0;
}
