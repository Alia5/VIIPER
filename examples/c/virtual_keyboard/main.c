#include "viiper/viiper.h"
#include "viiper/viiper_keyboard.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <time.h>

#if defined(_WIN32) || defined(_WIN64)
#  include <windows.h>
static void sleep_ms(int ms) { Sleep(ms); }
#else
#  include <unistd.h>
static void sleep_ms(int ms) { usleep(ms * 1000); }
#endif

static uint32_t choose_or_create_bus(viiper_client_t* client)
{
    viiper_bus_list_response_t list; memset(&list, 0, sizeof list);
    if (viiper_bus_list(client, &list) != VIIPER_OK) {
        fprintf(stderr, "BusList error: %s\n", viiper_get_error(client));
        return 0;
    }
    uint32_t busId = 0;
    if (list.buses_count > 0) {
        busId = list.Buses[0];
        for (size_t i = 1; i < list.buses_count; ++i)
            if (list.Buses[i] < busId) busId = list.Buses[i];
        printf("Using existing bus %u\n", (unsigned)busId);
        viiper_free_bus_list_response(&list);
        return busId;
    }
    viiper_free_bus_list_response(&list);

    viiper_bus_create_response_t cr; memset(&cr, 0, sizeof cr);
    if (viiper_bus_create(client, NULL, &cr) == VIIPER_OK) {
        printf("Created bus %u\n", (unsigned)cr.BusID);
        return cr.BusID;
    }
    fprintf(stderr, "BusCreate failed: %s\n", viiper_get_error(client));
    return 0;
}

static void on_leds(const void* output, size_t output_size, void* user)
{
    (void)user;
    if (!output || output_size == 0) return;
    const uint8_t* p = (const uint8_t*)output;
    /* Keyboard LED state is 1 byte messages; handle possible coalesced bytes */
    for (size_t i = 0; i < output_size; ++i) {
        uint8_t b = p[i];
        printf("→ LEDs: Num=%u Caps=%u Scroll=%u Compose=%u Kana=%u\n",
            (b & VIIPER_KEYBOARD_LEDNUMLOCK) ? 1u : 0u,
            (b & VIIPER_KEYBOARD_LEDCAPSLOCK) ? 1u : 0u,
            (b & VIIPER_KEYBOARD_LEDSCROLLLOCK) ? 1u : 0u,
            (b & VIIPER_KEYBOARD_LEDCOMPOSE) ? 1u : 0u,
            (b & VIIPER_KEYBOARD_LEDKANA) ? 1u : 0u);
    }
}

static void send_keys(viiper_device_t* dev, uint8_t modifiers, const uint8_t* keys, uint8_t nkeys)
{
    /* Build packet: [modifiers, count, keys...] with true N-key rollover */
    if (!keys && nkeys) return;
    /* Protocol uses u8 for count */
    size_t pkt_len = (size_t)2 + (size_t)nkeys;
    uint8_t* buf = (uint8_t*)malloc(pkt_len);
    if (!buf) return;
    buf[0] = modifiers;
    buf[1] = nkeys;
    if (nkeys && keys) memcpy(buf + 2, keys, nkeys);
    viiper_device_send(dev, buf, pkt_len);
    free(buf);
}

static void press_and_release(viiper_device_t* dev, uint8_t modifiers, uint8_t key)
{
    uint8_t keys[1] = { key };
    send_keys(dev, modifiers, keys, 1);
    sleep_ms(100);
    send_keys(dev, 0, NULL, 0); /* release all */
    sleep_ms(100);
}

static void type_string_hello(viiper_device_t* dev)
{
    /* H e l l o ! */
    press_and_release(dev, VIIPER_KEYBOARD_MODLEFTSHIFT, VIIPER_KEYBOARD_KEYH); /* 'H' */
    press_and_release(dev, 0, VIIPER_KEYBOARD_KEYE);
    press_and_release(dev, 0, VIIPER_KEYBOARD_KEYL);
    press_and_release(dev, 0, VIIPER_KEYBOARD_KEYL);
    press_and_release(dev, 0, VIIPER_KEYBOARD_KEYO);
    /* '!' = Shift + '1' */
    press_and_release(dev, VIIPER_KEYBOARD_MODLEFTSHIFT, VIIPER_KEYBOARD_KEY1);
}

int main(int argc, char** argv)
{
    /* Args: 0 -> default 127.0.0.1:3242; 1 -> host:port */
    char host[256]; uint16_t port = 3242;
    if (argc <= 1) {
        snprintf(host, sizeof host, "%s", "127.0.0.1");
    } else if (argc == 2) {
        const char* addr = argv[1]; const char* colon = strrchr(addr, ':');
        if (!colon) { fprintf(stderr, "Usage: virtual_keyboard [host:port]\n"); return 1; }
        size_t len = (size_t)(colon - addr); if (len >= sizeof(host)) len = sizeof(host)-1;
        memcpy(host, addr, len); host[len] = '\0';
        port = (uint16_t)atoi(colon + 1);
    } else { fprintf(stderr, "Usage: virtual_keyboard [host:port]\n"); return 1; }

    viiper_client_t* client = NULL;
    if (viiper_client_create(host, port, &client) != VIIPER_OK) {
        fprintf(stderr, "client_create failed\n");
        return 1;
    }

    uint32_t busId = choose_or_create_bus(client);
    if (busId == 0) { viiper_client_free(client); return 1; }

    /* Create and connect device in one step */
    viiper_device_info_t addResp; memset(&addResp, 0, sizeof addResp);
    viiper_device_t* dev = NULL;
    const char* deviceType = "keyboard";
    viiper_device_create_request_t createReq = {
        .Type = &deviceType,
        .IdVendor = NULL,
        .IdProduct = NULL
    };
    viiper_error_t rc = viiper_add_device_and_connect(client, busId, &createReq, &addResp, &dev);
    if (rc != VIIPER_OK) {
        fprintf(stderr, "add_device_and_connect failed: %s\n", viiper_get_error(client));
        viiper_client_free(client); return 1;
    }
    const char* devId = addResp.DevId ? addResp.DevId : "(none)";
    printf("Created and connected device %s on bus %u (type: %s)\n", devId, (unsigned)busId, addResp.Type ? addResp.Type : "unknown");

    /* Register LED callback */
    viiper_device_on_output(dev, on_leds, NULL);

    printf("Every 5s: type 'Hello!' + Enter. Press Ctrl+C to stop.\n");
    for (;;) {
        type_string_hello(dev);
        sleep_ms(100);
        press_and_release(dev, 0, VIIPER_KEYBOARD_KEYENTER);
        printf("→ Typed: Hello!\n");
        sleep_ms(5000);
    }

    viiper_device_close(dev);
    viiper_client_free(client);
    return 0;
}
