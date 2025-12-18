package cgen

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

const commonSourceTmpl = `/* Auto-generated VIIPER - C Client Library: common source */

#include "viiper.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#if defined(_WIN32) || defined(_WIN64)
#  include <winsock2.h>
#  include <ws2tcpip.h>
#  pragma comment(lib, "Ws2_32.lib")
static int viiper_wsa_init_once = 0;
#else
#  include <sys/types.h>
#  include <sys/socket.h>
#  include <netinet/tcp.h>
#  include <netdb.h>
#  include <unistd.h>
#  include <pthread.h>
#endif

/* ========================================================================
 * Internal structures
 * ======================================================================== */

struct viiper_client {
    int socket_fd;
    char* host;
    uint16_t port;
    char error_msg[256];
};

struct viiper_device {
    int socket_fd;
    viiper_client_t* client;
    uint32_t bus_id;
    char* dev_id;
    /* Async output handling */
    void* output_buffer;
    size_t output_buffer_size;
    viiper_output_cb callback;
    void* callback_user;
    /* Disconnect callback */
    viiper_disconnect_cb disconnect_callback;
    void* disconnect_user;
    int running;
#if defined(_WIN32) || defined(_WIN64)
    HANDLE recv_thread;
#else
    pthread_t recv_thread;
#endif
};

/* ========================================================================
 * Minimal JSON helpers (sufficient for VIIPER API responses)
 * ======================================================================== */

static const char* json_skip_ws(const char* s){ while(*s==' '||*s=='\t'||*s=='\r'||*s=='\n') ++s; return s; }
static int json_streq(const char* a,const char* b){ return strcmp(a,b)==0; }

static const char* json_find_key(const char* json, const char* key){
    const size_t klen = strlen(key);
    const char* p = json;
    while ((p = strchr(p, '"')) != NULL) {
        const char* q = strchr(p+1, '"'); if (!q) return NULL;
        size_t len = (size_t)(q - (p+1));
        if (len == klen && strncmp(p+1, key, klen) == 0) {
            const char* c = json_skip_ws(q+1);
            if (*c == ':') return json_skip_ws(c+1);
        }
        p = q+1;
    }
    return NULL;
}

static int json_parse_uint32(const char* json, const char* key, uint32_t* out){
    const char* v = json_find_key(json, key); if (!v) return -1;
    unsigned long val = 0; int neg=0;
    if (*v=='-'){ neg=1; ++v; }
    while (*v>='0' && *v<='9'){ val = val*10 + (unsigned long)(*v - '0'); ++v; }
    if (neg) return -1; *out = (uint32_t)val; return 0;
}

static int json_parse_string_alloc(const char* json, const char* key, char** out){
    const char* v = json_find_key(json, key); if (!v) return -1;
    if (*v!='"') return -1; ++v; const char* q = v; while (*q && *q!='"') ++q; if (*q!='"') return -1;
    size_t n = (size_t)(q - v);
    char* s = (char*)malloc(n+1); if (!s) return -1; memcpy(s, v, n); s[n] = '\0'; *out = s; return 0;
}

static int json_parse_array_uint32(const char* json, const char* key, uint32_t** arr, size_t* count){
    const char* v = json_find_key(json, key); if (!v) return -1;
    if (*v!='[') return -1; ++v;
    size_t cap=8, n=0; uint32_t* out = (uint32_t*)malloc(cap*sizeof(uint32_t)); if(!out) return -1;
    v = json_skip_ws(v);
    while (*v && *v!=']'){
        unsigned long val=0; int got=0;
        while (*v>='0' && *v<='9'){ val = val*10 + (unsigned long)(*v - '0'); ++v; got=1; }
        if (!got){ free(out); return -1; }
        if (n>=cap){ cap*=2; uint32_t* tmp=(uint32_t*)realloc(out,cap*sizeof(uint32_t)); if(!tmp){ free(out); return -1; } out=tmp; }
        out[n++] = (uint32_t)val;
        v = json_skip_ws(v);
        if (*v==','){ ++v; v = json_skip_ws(v); }
    }
    if (*v!=']'){ free(out); return -1; }
    *arr = out; *count = n; return 0;
}

/* Device info object parser for arrays */
static int json_parse_device_info_obj(const char* obj, viiper_device_info_t* out){
    if (json_parse_uint32(obj, "busId", &out->BusID) != 0) return -1;
    if (json_parse_string_alloc(obj, "devId", (char**)&out->DevId) != 0) return -1;
    if (json_parse_string_alloc(obj, "vid", (char**)&out->Vid) != 0) return -1;
    if (json_parse_string_alloc(obj, "pid", (char**)&out->Pid) != 0) return -1;
    if (json_parse_string_alloc(obj, "type", (char**)&out->Type) != 0) return -1;
    return 0;
}

static int json_parse_array_device_info(const char* json, const char* key, viiper_device_info_t** arr, size_t* count){
    const char* v = json_find_key(json, key); if (!v) return -1;
    if (*v!='[') return -1; ++v;
    size_t cap=4, n=0; viiper_device_info_t* out = (viiper_device_info_t*)calloc(cap, sizeof(viiper_device_info_t)); if(!out) return -1;
    v = json_skip_ws(v);
    while (*v && *v!=']'){
        if (*v!='{'){ free(out); return -1; }
        int depth=1; const char* start=v; ++v; while (*v && depth>0){ if (*v=='{') depth++; else if (*v=='}') depth--; ++v; }
        if (depth!=0){ free(out); return -1; }
        const char* end = v; // points after closing '}'
        size_t len = (size_t)(end - start);
        char* slice = (char*)malloc(len+1); if(!slice){ free(out); return -1; }
        memcpy(slice, start, len); slice[len]='\0';
        if (n>=cap){ cap*=2; viiper_device_info_t* tmp=(viiper_device_info_t*)realloc(out,cap*sizeof(viiper_device_info_t)); if(!tmp){ free(slice); free(out); return -1; } out=tmp; }
        if (json_parse_device_info_obj(slice, &out[n]) != 0){ free(slice); free(out); return -1; }
        free(slice);
        n++;
        v = json_skip_ws(v);
        if (*v==','){ ++v; v = json_skip_ws(v); }
    }
    if (*v!=']'){ free(out); return -1; }
    *arr = out; *count = n; return 0;
}

/* ========================================================================
 * Client API
 * ======================================================================== */

VIIPER_API viiper_error_t viiper_client_create(
    const char* host,
    uint16_t port,
    viiper_client_t** out_client
) {
    viiper_client_t* client = (viiper_client_t*)calloc(1, sizeof(viiper_client_t));
    if (!client) {
        return VIIPER_ERROR_MEMORY;
    }
    if (host) {
        size_t len = strlen(host);
        client->host = (char*)malloc(len + 1);
        if (!client->host) {
            free(client);
            return VIIPER_ERROR_MEMORY;
        }
        memcpy(client->host, host, len + 1);
    } else {
        client->host = NULL;
    }
    client->port = port;
    client->socket_fd = -1;
    *out_client = client;
    return VIIPER_OK;
}

VIIPER_API void viiper_client_free(viiper_client_t* client) {
    if (!client) return;
    if (client->host) free(client->host);
    free(client);
}

VIIPER_API const char* viiper_get_error(viiper_client_t* client) {
    if (!client) return "Invalid client";
    return client->error_msg;
}

/* ========================================================================
 * Internal networking helpers
 * ======================================================================== */

static int viiper_connect(const char* host, uint16_t port) {
#if defined(_WIN32) || defined(_WIN64)
    if (!viiper_wsa_init_once) {
        WSADATA wsaData;
        if (WSAStartup(MAKEWORD(2,2), &wsaData) != 0) {
            return -1;
        }
        viiper_wsa_init_once = 1;
    }
#endif
    char portbuf[16];
    snprintf(portbuf, sizeof portbuf, "%u", (unsigned)port);
    struct addrinfo hints; memset(&hints, 0, sizeof hints);
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    hints.ai_protocol = IPPROTO_TCP;
    struct addrinfo* res = NULL;
    if (getaddrinfo(host, portbuf, &hints, &res) != 0 || !res) {
        return -1;
    }
    int fd = -1;
    for (struct addrinfo* p = res; p; p = p->ai_next) {
        int s = (int)socket(p->ai_family, p->ai_socktype, p->ai_protocol);
        if (s < 0) continue;
        if (connect(s, p->ai_addr, (int)p->ai_addrlen) == 0) {
            int flag = 1;
#if defined(_WIN32) || defined(_WIN64)
            setsockopt(s, IPPROTO_TCP, TCP_NODELAY, (const char*)&flag, sizeof(flag));
#else
            setsockopt(s, IPPROTO_TCP, TCP_NODELAY, &flag, sizeof(flag));
#endif
            fd = s;
            break;
        }
#if defined(_WIN32) || defined(_WIN64)
        closesocket(s);
#else
        close(s);
#endif
    }
    freeaddrinfo(res);
    return fd;
}

static int viiper_send_line(int fd, const char* line) {
    size_t n = strlen(line);
#if defined(_WIN32) || defined(_WIN64)
    int wr = send(fd, line, (int)n, 0);
    if (wr < 0) return -1;
    wr = send(fd, "\0", 1, 0);
    if (wr < 0) return -1;
#else
    ssize_t wr = send(fd, line, n, 0);
    if (wr < 0) return -1;
    wr = send(fd, "\0", 1, 0);
    if (wr < 0) return -1;
#endif
    return 0;
}

static int viiper_read_line(int fd, char** out) {
    size_t cap = 256, len = 0;
    char* buf = (char*)malloc(cap);
    if (!buf) return -1;
    for (;;) {
        char ch;
#if defined(_WIN32) || defined(_WIN64)
        int rd = recv(fd, &ch, 1, 0);
#else
        ssize_t rd = recv(fd, &ch, 1, 0);
#endif
        if (rd < 0) { free(buf); return -1; }
        if (rd == 0) break;
        if (ch == '\0') break;
        if (len + 1 >= cap) {
            cap *= 2;
            char* nb = (char*)realloc(buf, cap);
            if (!nb) { free(buf); return -1; }
            buf = nb;
        }
        buf[len++] = ch;
    }
    buf[len] = '\0';
    *out = buf;
    return 0;
}

static int viiper_do(viiper_client_t* client, const char* path, const char* payload, char** out_line) {
    if (!client || !client->host || client->port == 0) return -1;
    int fd = viiper_connect(client->host, client->port);
    if (fd < 0) return -1;
    size_t need = strlen(path) + 1 + (payload ? strlen(payload) + 1 : 0);
    char* line = (char*)malloc(need);
    if (!line) {
#if defined(_WIN32) || defined(_WIN64)
        closesocket(fd);
#else
        close(fd);
#endif
        return -1;
    }
    if (payload && payload[0]) {
        snprintf(line, need, "%s %s", path, payload);
    } else {
        snprintf(line, need, "%s", path);
    }
    int rc = viiper_send_line(fd, line);
    free(line);
    if (rc != 0) {
#if defined(_WIN32) || defined(_WIN64)
        closesocket(fd);
#else
        close(fd);
#endif
        return -1;
    }
    char* resp = NULL;
    rc = viiper_read_line(fd, &resp);
#if defined(_WIN32) || defined(_WIN64)
    closesocket(fd);
#else
    close(fd);
#endif
    if (rc != 0) return -1;
    *out_line = resp;
    return 0;
}

/* ========================================================================
 * DTO parsers and free helpers
 * ======================================================================== */

/* Free helpers */
{{range .DTOs}}{{if ne .Name "Device"}}
{{genFreeFunc .}}
{{end}}{{end}}

/* Parsers */
static int viiper_parse_device_info_obj(const char* json, viiper_device_info_t* out){ if (json_parse_uint32(json, "busId", &out->BusID)!=0) return -1; if (json_parse_string_alloc(json, "devId", (char**)&out->DevId)!=0) return -1; json_parse_string_alloc(json, "vid", (char**)&out->Vid); json_parse_string_alloc(json, "pid", (char**)&out->Pid); json_parse_string_alloc(json, "type", (char**)&out->Type); return 0; }
{{range .DTOs}}{{if ne .Name "Device"}}
{{genParser .}}
{{end}}{{end}}

/* ========================================================================
 * Management API - Implementations
 * ======================================================================== */
{{range .Routes}}
VIIPER_API viiper_error_t viiper_{{snakecase .Handler}}(
    viiper_client_t* client{{ $params := pathParams .Path }}{{range $params}},
    const char* {{.}}{{end}}{{$payloadType := payloadCType .Payload}}{{if ne $payloadType ""}},
    {{$payloadType}} {{if eq .Payload.Kind "json"}}request{{else if eq .Payload.Kind "numeric"}}payload_value{{else}}payload_str{{end}}{{end}}{{if .ResponseDTO}},
    {{responseCType .ResponseDTO}}* out{{end}}
) {
    if (!client) return VIIPER_ERROR_INVALID_PARAM;
    /* Build path by substituting params in order */
    char pathbuf[256];
    const char* pattern = "{{.Path}}";
    snprintf(pathbuf, sizeof pathbuf, "%s", pattern);
    {{ $params := pathParams .Path }}
    {{range $i, $p := $params}}
    { /* replace {{$p}} placeholder with provided argument */
        const char* needle = "{{printf "{%s}" $p}}";
        char tmp[256]; tmp[0]='\0';
        char* pos = strstr(pathbuf, needle);
        if (pos) {
            size_t head = (size_t)(pos - pathbuf);
            snprintf(tmp, sizeof tmp, "%.*s%s%s", (int)head, pathbuf, {{ $p }}, pos + strlen(needle));
            snprintf(pathbuf, sizeof pathbuf, "%s", tmp);
        }
    }
    {{end}}
    /* Build payload based on PayloadKind */
    char payload[512]; payload[0]='\0';
    {{if eq .Payload.Kind "json"}}
    {{marshalPayload .Payload}}
    {{else if eq .Payload.Kind "numeric"}}{{if .Payload.Required}}
    snprintf(payload, sizeof payload, "%u", (unsigned)payload_value);
    {{else}}
    if (payload_value) {
        snprintf(payload, sizeof payload, "%u", (unsigned)*payload_value);
    }
    {{end}}
    {{else if eq .Payload.Kind "string"}}
    if (payload_str && payload_str[0]) {
        snprintf(payload, sizeof payload, "%s", payload_str);
    }
    {{end}}
    char* line = NULL;
    if (viiper_do(client, pathbuf, payload[0]?payload:NULL, &line) != 0) {
        snprintf(client->error_msg, sizeof client->error_msg, "io error");
        return VIIPER_ERROR_IO;
    }
    /* rudimentary error detection: {"status":4xx/5xx} (RFC 7807) */
    if (line && strncmp(line, "{\"status\":", 11) == 0) {
        snprintf(client->error_msg, sizeof client->error_msg, "%s", line);
        free(line);
        return VIIPER_ERROR_PROTOCOL;
    }
    {{if .ResponseDTO}}
    if (out) {
        int prc = 0;
        {{if eq .ResponseDTO "Device"}}prc = viiper_parse_device_info_obj(line, out);{{else}}prc = viiper_parse_{{snakecase .ResponseDTO}}(line, out);{{end}}
        if (prc != 0) { snprintf(client->error_msg, sizeof client->error_msg, "parse error"); free(line); return VIIPER_ERROR_PROTOCOL; }
    }
    free(line);
    {{else}}
    free(line);
    {{end}}
    return VIIPER_OK;
}
{{end}}

/* ========================================================================
 * Device Streaming API Implementation
 * ======================================================================== */

#if defined(_WIN32) || defined(_WIN64)
static DWORD WINAPI viiper_device_receiver_thread(LPVOID arg) {
#else
static void* viiper_device_receiver_thread(void* arg) {
#endif
    viiper_device_t* dev = (viiper_device_t*)arg;
    while (dev->running) {
#if defined(_WIN32) || defined(_WIN64)
        DWORD timeout_ms = 200;
        fd_set rfds; FD_ZERO(&rfds); FD_SET((SOCKET)dev->socket_fd, &rfds);
        struct timeval tv; tv.tv_sec = 0; tv.tv_usec = timeout_ms * 1000;
        int sel = select(0, &rfds, NULL, NULL, &tv);
        if (sel < 0) break;
        if (sel == 0) continue; /* timeout */
        if (dev->callback && dev->output_buffer && dev->output_buffer_size > 0) {
            int rd = recv(dev->socket_fd, (char*)dev->output_buffer, (int)dev->output_buffer_size, 0);
            if (rd <= 0) break;
            dev->callback(dev->output_buffer, (size_t)rd, dev->callback_user);
        }
#else
        fd_set rfds; FD_ZERO(&rfds); FD_SET(dev->socket_fd, &rfds);
        struct timeval tv; tv.tv_sec = 0; tv.tv_usec = 200000;
        int sel = select(dev->socket_fd+1, &rfds, NULL, NULL, &tv);
        if (sel < 0) break;
        if (sel == 0) continue; /* timeout */
        if (dev->callback && dev->output_buffer && dev->output_buffer_size > 0) {
            ssize_t rd = recv(dev->socket_fd, dev->output_buffer, dev->output_buffer_size, 0);
            if (rd <= 0) break;
            dev->callback(dev->output_buffer, (size_t)rd, dev->callback_user);
        }
#endif
    }
    /* Call disconnect callback if registered */
    if (dev->disconnect_callback) {
        dev->disconnect_callback(dev->disconnect_user);
    }
#if defined(_WIN32) || defined(_WIN64)
    return 0;
#else
    return NULL;
#endif
}

VIIPER_API viiper_error_t viiper_device_create(
    viiper_client_t* client,
    uint32_t bus_id,
    const char* dev_id,
    viiper_device_t** out_device
) {
    if (!client || !dev_id || !out_device) return VIIPER_ERROR_INVALID_PARAM;
    int fd = viiper_connect(client->host, client->port);
    if (fd < 0) return VIIPER_ERROR_CONNECT;
    /* Send stream path with null terminator for framing */
    char pathbuf[256];
    snprintf(pathbuf, sizeof pathbuf, "bus/%u/%s", (unsigned)bus_id, dev_id);
    if (viiper_send_line(fd, pathbuf) != 0) {
#if defined(_WIN32) || defined(_WIN64)
        closesocket(fd);
#else
        close(fd);
#endif
        return VIIPER_ERROR_IO;
    }
    viiper_device_t* dev = (viiper_device_t*)calloc(1, sizeof(viiper_device_t));
    if (!dev) {
#if defined(_WIN32) || defined(_WIN64)
        closesocket(fd);
#else
        close(fd);
#endif
        return VIIPER_ERROR_MEMORY;
    }
    dev->socket_fd = fd;
    dev->client = client;
    dev->bus_id = bus_id;
    size_t idlen = strlen(dev_id);
    dev->dev_id = (char*)malloc(idlen+1);
    if (!dev->dev_id) {
#if defined(_WIN32) || defined(_WIN64)
        closesocket(fd);
#else
        close(fd);
#endif
        free(dev);
        return VIIPER_ERROR_MEMORY;
    }
    memcpy(dev->dev_id, dev_id, idlen+1);
    dev->running = 1;
    /* Start receiver thread */
#if defined(_WIN32) || defined(_WIN64)
    dev->recv_thread = CreateThread(NULL, 0, viiper_device_receiver_thread, dev, 0, NULL);
    if (!dev->recv_thread) {
        closesocket(fd);
        free(dev->dev_id);
        free(dev);
        return VIIPER_ERROR_IO;
    }
#else
    if (pthread_create(&dev->recv_thread, NULL, viiper_device_receiver_thread, dev) != 0) {
        close(fd);
        free(dev->dev_id);
        free(dev);
        return VIIPER_ERROR_IO;
    }
#endif
    *out_device = dev;
    return VIIPER_OK;
}

VIIPER_API viiper_error_t viiper_device_send(
    viiper_device_t* device,
    const void* input,
    size_t input_size
) {
    if (!device || !input) return VIIPER_ERROR_INVALID_PARAM;
#if defined(_WIN32) || defined(_WIN64)
    int wr = send(device->socket_fd, (const char*)input, (int)input_size, 0);
#else
    ssize_t wr = send(device->socket_fd, input, input_size, 0);
#endif
    return (wr < 0) ? VIIPER_ERROR_IO : VIIPER_OK;
}

VIIPER_API void viiper_device_on_output(
    viiper_device_t* device,
    void* buffer,
    size_t buffer_size,
    viiper_output_cb callback,
    void* user_data
) {
    if (!device) return;
    device->output_buffer = buffer;
    device->output_buffer_size = buffer_size;
    device->callback = callback;
    device->callback_user = user_data;
}

VIIPER_API void viiper_device_on_disconnect(
    viiper_device_t* device,
    viiper_disconnect_cb callback,
    void* user_data
) {
    if (!device) return;
    device->disconnect_callback = callback;
    device->disconnect_user = user_data;
}

VIIPER_API void viiper_device_close(viiper_device_t* device) {
    if (!device) return;
    device->running = 0;
#if defined(_WIN32) || defined(_WIN64)
    shutdown(device->socket_fd, SD_BOTH);
    if (device->recv_thread) {
        WaitForSingleObject(device->recv_thread, INFINITE);
        CloseHandle(device->recv_thread);
    }
    closesocket(device->socket_fd);
#else
    shutdown(device->socket_fd, SHUT_RDWR);
    pthread_join(device->recv_thread, NULL);
    close(device->socket_fd);
#endif
    if (device->dev_id) free(device->dev_id);
    free(device);
}

/* OpenStream: connect to an existing device's stream channel (device must already exist on bus) */
VIIPER_API viiper_error_t viiper_open_stream(
    viiper_client_t* client,
    uint32_t bus_id,
    const char* dev_id,
    viiper_device_t** out_device
) {
    if (!client || !dev_id || !out_device) return VIIPER_ERROR_INVALID_PARAM;
    int fd = viiper_connect(client->host, client->port);
    if (fd < 0) return VIIPER_ERROR_CONNECT;
    char pathbuf[256];
    snprintf(pathbuf, sizeof pathbuf, "bus/%u/%s", (unsigned)bus_id, dev_id);
    if (viiper_send_line(fd, pathbuf) != 0) {
#if defined(_WIN32) || defined(_WIN64)
        closesocket(fd);
#else
        close(fd);
#endif
        return VIIPER_ERROR_IO;
    }
    viiper_device_t* dev = (viiper_device_t*)calloc(1, sizeof(viiper_device_t));
    if (!dev) {
#if defined(_WIN32) || defined(_WIN64)
        closesocket(fd);
#else
        close(fd);
#endif
        return VIIPER_ERROR_MEMORY;
    }
    dev->socket_fd = fd;
    dev->client = client;
    dev->bus_id = bus_id;
    size_t idlen = strlen(dev_id);
    dev->dev_id = (char*)malloc(idlen+1);
    if (!dev->dev_id) {
#if defined(_WIN32) || defined(_WIN64)
        closesocket(fd);
#else
        close(fd);
#endif
        free(dev);
        return VIIPER_ERROR_MEMORY;
    }
    memcpy(dev->dev_id, dev_id, idlen+1);
    dev->running = 1;
#if defined(_WIN32) || defined(_WIN64)
    dev->recv_thread = CreateThread(NULL, 0, viiper_device_receiver_thread, dev, 0, NULL);
    if (!dev->recv_thread) {
        closesocket(fd); free(dev->dev_id); free(dev); return VIIPER_ERROR_IO;
    }
#else
    if (pthread_create(&dev->recv_thread, NULL, viiper_device_receiver_thread, dev) != 0) {
        close(fd); free(dev->dev_id); free(dev); return VIIPER_ERROR_IO;
    }
#endif
    *out_device = dev;
    return VIIPER_OK;
}

/* Convenience wrapper: AddDeviceAndConnect (create device on bus then open stream) */
VIIPER_API viiper_error_t viiper_add_device_and_connect(
    viiper_client_t* client,
    uint32_t bus_id,
    const viiper_device_create_request_t* request,
    viiper_device_info_t* out_info,
    viiper_device_t** out_device
) {
    if (!client || !out_info || !out_device) return VIIPER_ERROR_INVALID_PARAM;
    char busIdStr[32]; snprintf(busIdStr, sizeof busIdStr, "%u", (unsigned)bus_id);
    viiper_error_t rc = viiper_bus_device_add(client, busIdStr, request, out_info);
    if (rc != VIIPER_OK) return rc;
    const char* devId = out_info->DevId ? out_info->DevId : NULL;
    if (!devId) return VIIPER_ERROR_PROTOCOL; /* missing devId */
    return viiper_open_stream(client, bus_id, devId, out_device);
}
`

func generateCommonSource(logger *slog.Logger, srcDir string, md *meta.Metadata) error {
	out := filepath.Join(srcDir, "viiper.c")
	t := template.Must(template.New("viiper.c").Funcs(tplFuncs(md)).Parse(commonSourceTmpl))
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create source: %w", err)
	}
	defer f.Close()
	if err := t.Execute(f, md); err != nil {
		return fmt.Errorf("exec source tmpl: %w", err)
	}
	logger.Info("Generated common source", "file", out)
	return nil
}
