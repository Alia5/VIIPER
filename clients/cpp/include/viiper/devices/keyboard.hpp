

#pragma once

#include "../error.hpp"
#include "../detail/json.hpp"
#include <cstdint>
#include <vector>
#include <array>
#include <string_view>
#include <algorithm>
#include <optional>
#include <unordered_map>
#include <unordered_set>

namespace viiper
{
    namespace keyboard
    {

        // ============================================================================
        // Constants
        // ============================================================================

        constexpr std::size_t OUTPUT_SIZE = 1;

        constexpr std::uint64_t ModLeftCtrl = 1;
        constexpr std::uint64_t ModLeftShift = 2;
        constexpr std::uint64_t ModLeftAlt = 4;
        constexpr std::uint64_t ModLeftGUI = 8;
        constexpr std::uint64_t ModRightCtrl = 16;
        constexpr std::uint64_t ModRightShift = 32;
        constexpr std::uint64_t ModRightAlt = 64;
        constexpr std::uint64_t ModRightGUI = 128;
        constexpr std::uint64_t LEDNumLock = 1;
        constexpr std::uint64_t LEDCapsLock = 2;
        constexpr std::uint64_t LEDScrollLock = 4;
        constexpr std::uint64_t LEDCompose = 8;
        constexpr std::uint64_t LEDKana = 16;
        constexpr std::uint64_t KeyA = 4;
        constexpr std::uint64_t KeyB = 5;
        constexpr std::uint64_t KeyC = 6;
        constexpr std::uint64_t KeyD = 7;
        constexpr std::uint64_t KeyE = 8;
        constexpr std::uint64_t KeyF = 9;
        constexpr std::uint64_t KeyG = 10;
        constexpr std::uint64_t KeyH = 11;
        constexpr std::uint64_t KeyI = 12;
        constexpr std::uint64_t KeyJ = 13;
        constexpr std::uint64_t KeyK = 14;
        constexpr std::uint64_t KeyL = 15;
        constexpr std::uint64_t KeyM = 16;
        constexpr std::uint64_t KeyN = 17;
        constexpr std::uint64_t KeyO = 18;
        constexpr std::uint64_t KeyP = 19;
        constexpr std::uint64_t KeyQ = 20;
        constexpr std::uint64_t KeyR = 21;
        constexpr std::uint64_t KeyS = 22;
        constexpr std::uint64_t KeyT = 23;
        constexpr std::uint64_t KeyU = 24;
        constexpr std::uint64_t KeyV = 25;
        constexpr std::uint64_t KeyW = 26;
        constexpr std::uint64_t KeyX = 27;
        constexpr std::uint64_t KeyY = 28;
        constexpr std::uint64_t KeyZ = 29;
        constexpr std::uint64_t Key1 = 30;
        constexpr std::uint64_t Key2 = 31;
        constexpr std::uint64_t Key3 = 32;
        constexpr std::uint64_t Key4 = 33;
        constexpr std::uint64_t Key5 = 34;
        constexpr std::uint64_t Key6 = 35;
        constexpr std::uint64_t Key7 = 36;
        constexpr std::uint64_t Key8 = 37;
        constexpr std::uint64_t Key9 = 38;
        constexpr std::uint64_t Key0 = 39;
        constexpr std::uint64_t KeyEnter = 40;
        constexpr std::uint64_t KeyEscape = 41;
        constexpr std::uint64_t KeyBackspace = 42;
        constexpr std::uint64_t KeyTab = 43;
        constexpr std::uint64_t KeySpace = 44;
        constexpr std::uint64_t KeyMinus = 45;
        constexpr std::uint64_t KeyEqual = 46;
        constexpr std::uint64_t KeyLeftBrace = 47;
        constexpr std::uint64_t KeyRightBrace = 48;
        constexpr std::uint64_t KeyBackslash = 49;
        constexpr std::uint64_t KeyNonUSHash = 50;
        constexpr std::uint64_t KeySemicolon = 51;
        constexpr std::uint64_t KeyApostrophe = 52;
        constexpr std::uint64_t KeyGrave = 53;
        constexpr std::uint64_t KeyComma = 54;
        constexpr std::uint64_t KeyPeriod = 55;
        constexpr std::uint64_t KeySlash = 56;
        constexpr std::uint64_t KeyCapsLock = 57;
        constexpr std::uint64_t KeyF1 = 58;
        constexpr std::uint64_t KeyF2 = 59;
        constexpr std::uint64_t KeyF3 = 60;
        constexpr std::uint64_t KeyF4 = 61;
        constexpr std::uint64_t KeyF5 = 62;
        constexpr std::uint64_t KeyF6 = 63;
        constexpr std::uint64_t KeyF7 = 64;
        constexpr std::uint64_t KeyF8 = 65;
        constexpr std::uint64_t KeyF9 = 66;
        constexpr std::uint64_t KeyF10 = 67;
        constexpr std::uint64_t KeyF11 = 68;
        constexpr std::uint64_t KeyF12 = 69;
        constexpr std::uint64_t KeyPrintScreen = 70;
        constexpr std::uint64_t KeyScrollLock = 71;
        constexpr std::uint64_t KeyPause = 72;
        constexpr std::uint64_t KeyInsert = 73;
        constexpr std::uint64_t KeyHome = 74;
        constexpr std::uint64_t KeyPageUp = 75;
        constexpr std::uint64_t KeyDelete = 76;
        constexpr std::uint64_t KeyEnd = 77;
        constexpr std::uint64_t KeyPageDown = 78;
        constexpr std::uint64_t KeyRight = 79;
        constexpr std::uint64_t KeyLeft = 80;
        constexpr std::uint64_t KeyDown = 81;
        constexpr std::uint64_t KeyUp = 82;
        constexpr std::uint64_t KeyNumLock = 83;
        constexpr std::uint64_t KeyKpSlash = 84;
        constexpr std::uint64_t KeyKpAsterisk = 85;
        constexpr std::uint64_t KeyKpMinus = 86;
        constexpr std::uint64_t KeyKpPlus = 87;
        constexpr std::uint64_t KeyKpEnter = 88;
        constexpr std::uint64_t KeyKp1 = 89;
        constexpr std::uint64_t KeyKp2 = 90;
        constexpr std::uint64_t KeyKp3 = 91;
        constexpr std::uint64_t KeyKp4 = 92;
        constexpr std::uint64_t KeyKp5 = 93;
        constexpr std::uint64_t KeyKp6 = 94;
        constexpr std::uint64_t KeyKp7 = 95;
        constexpr std::uint64_t KeyKp8 = 96;
        constexpr std::uint64_t KeyKp9 = 97;
        constexpr std::uint64_t KeyKp0 = 98;
        constexpr std::uint64_t KeyKpDot = 99;
        constexpr std::uint64_t KeyNonUSBackslash = 100;
        constexpr std::uint64_t KeyApplication = 101;
        constexpr std::uint64_t KeyPower = 102;
        constexpr std::uint64_t KeyKpEqual = 103;
        constexpr std::uint64_t KeyF13 = 104;
        constexpr std::uint64_t KeyF14 = 105;
        constexpr std::uint64_t KeyF15 = 106;
        constexpr std::uint64_t KeyF16 = 107;
        constexpr std::uint64_t KeyF17 = 108;
        constexpr std::uint64_t KeyF18 = 109;
        constexpr std::uint64_t KeyF19 = 110;
        constexpr std::uint64_t KeyF20 = 111;
        constexpr std::uint64_t KeyF21 = 112;
        constexpr std::uint64_t KeyF22 = 113;
        constexpr std::uint64_t KeyF23 = 114;
        constexpr std::uint64_t KeyF24 = 115;
        constexpr std::uint64_t KeyExecute = 116;
        constexpr std::uint64_t KeyHelp = 117;
        constexpr std::uint64_t KeyMenu = 118;
        constexpr std::uint64_t KeySelect = 119;
        constexpr std::uint64_t KeyStop = 120;
        constexpr std::uint64_t KeyAgain = 121;
        constexpr std::uint64_t KeyUndo = 122;
        constexpr std::uint64_t KeyCut = 123;
        constexpr std::uint64_t KeyCopy = 124;
        constexpr std::uint64_t KeyPaste = 125;
        constexpr std::uint64_t KeyFind = 126;
        constexpr std::uint64_t KeyMute = 127;
        constexpr std::uint64_t KeyVolumeUp = 128;
        constexpr std::uint64_t KeyVolumeDown = 129;
        constexpr std::uint64_t KeyMediaPlayPause = 232;
        constexpr std::uint64_t KeyMediaStop = 233;
        constexpr std::uint64_t KeyMediaNext = 235;
        constexpr std::uint64_t KeyMediaPrevious = 236;

        inline constexpr std::array<std::pair<std::uint64_t, std::string_view>, 115> KEY_NAME = {{{Key0, "0"},
                                                                                                  {Key1, "1"},
                                                                                                  {Key2, "2"},
                                                                                                  {Key3, "3"},
                                                                                                  {Key4, "4"},
                                                                                                  {Key5, "5"},
                                                                                                  {Key6, "6"},
                                                                                                  {Key7, "7"},
                                                                                                  {Key8, "8"},
                                                                                                  {Key9, "9"},
                                                                                                  {KeyA, "A"},
                                                                                                  {KeyApostrophe, "Apostrophe"},
                                                                                                  {KeyApplication, "Application"},
                                                                                                  {KeyB, "B"},
                                                                                                  {KeyBackslash, "Backslash"},
                                                                                                  {KeyBackspace, "Backspace"},
                                                                                                  {KeyC, "C"},
                                                                                                  {KeyCapsLock, "CapsLock"},
                                                                                                  {KeyComma, "Comma"},
                                                                                                  {KeyD, "D"},
                                                                                                  {KeyDelete, "Delete"},
                                                                                                  {KeyDown, "Down"},
                                                                                                  {KeyE, "E"},
                                                                                                  {KeyEnd, "End"},
                                                                                                  {KeyEnter, "Enter"},
                                                                                                  {KeyEqual, "Equal"},
                                                                                                  {KeyEscape, "Escape"},
                                                                                                  {KeyF, "F"},
                                                                                                  {KeyF1, "F1"},
                                                                                                  {KeyF10, "F10"},
                                                                                                  {KeyF11, "F11"},
                                                                                                  {KeyF12, "F12"},
                                                                                                  {KeyF13, "F13"},
                                                                                                  {KeyF14, "F14"},
                                                                                                  {KeyF15, "F15"},
                                                                                                  {KeyF16, "F16"},
                                                                                                  {KeyF17, "F17"},
                                                                                                  {KeyF18, "F18"},
                                                                                                  {KeyF19, "F19"},
                                                                                                  {KeyF2, "F2"},
                                                                                                  {KeyF20, "F20"},
                                                                                                  {KeyF21, "F21"},
                                                                                                  {KeyF22, "F22"},
                                                                                                  {KeyF23, "F23"},
                                                                                                  {KeyF24, "F24"},
                                                                                                  {KeyF3, "F3"},
                                                                                                  {KeyF4, "F4"},
                                                                                                  {KeyF5, "F5"},
                                                                                                  {KeyF6, "F6"},
                                                                                                  {KeyF7, "F7"},
                                                                                                  {KeyF8, "F8"},
                                                                                                  {KeyF9, "F9"},
                                                                                                  {KeyG, "G"},
                                                                                                  {KeyGrave, "Grave"},
                                                                                                  {KeyH, "H"},
                                                                                                  {KeyHome, "Home"},
                                                                                                  {KeyI, "I"},
                                                                                                  {KeyInsert, "Insert"},
                                                                                                  {KeyJ, "J"},
                                                                                                  {KeyK, "K"},
                                                                                                  {KeyKp0, "Kp0"},
                                                                                                  {KeyKp1, "Kp1"},
                                                                                                  {KeyKp2, "Kp2"},
                                                                                                  {KeyKp3, "Kp3"},
                                                                                                  {KeyKp4, "Kp4"},
                                                                                                  {KeyKp5, "Kp5"},
                                                                                                  {KeyKp6, "Kp6"},
                                                                                                  {KeyKp7, "Kp7"},
                                                                                                  {KeyKp8, "Kp8"},
                                                                                                  {KeyKp9, "Kp9"},
                                                                                                  {KeyKpAsterisk, "Kp*"},
                                                                                                  {KeyKpDot, "Kp."},
                                                                                                  {KeyKpEnter, "KpEnter"},
                                                                                                  {KeyKpMinus, "Kp-"},
                                                                                                  {KeyKpPlus, "Kp+"},
                                                                                                  {KeyKpSlash, "Kp/"},
                                                                                                  {KeyL, "L"},
                                                                                                  {KeyLeft, "Left"},
                                                                                                  {KeyLeftBrace, "LeftBrace"},
                                                                                                  {KeyM, "M"},
                                                                                                  {KeyMediaNext, "MediaNext"},
                                                                                                  {KeyMediaPlayPause, "MediaPlayPause"},
                                                                                                  {KeyMediaPrevious, "MediaPrevious"},
                                                                                                  {KeyMediaStop, "MediaStop"},
                                                                                                  {KeyMinus, "Minus"},
                                                                                                  {KeyMute, "Mute"},
                                                                                                  {KeyN, "N"},
                                                                                                  {KeyNumLock, "NumLock"},
                                                                                                  {KeyO, "O"},
                                                                                                  {KeyP, "P"},
                                                                                                  {KeyPageDown, "PageDown"},
                                                                                                  {KeyPageUp, "PageUp"},
                                                                                                  {KeyPause, "Pause"},
                                                                                                  {KeyPeriod, "Period"},
                                                                                                  {KeyPrintScreen, "PrintScreen"},
                                                                                                  {KeyQ, "Q"},
                                                                                                  {KeyR, "R"},
                                                                                                  {KeyRight, "Right"},
                                                                                                  {KeyRightBrace, "RightBrace"},
                                                                                                  {KeyS, "S"},
                                                                                                  {KeyScrollLock, "ScrollLock"},
                                                                                                  {KeySemicolon, "Semicolon"},
                                                                                                  {KeySlash, "Slash"},
                                                                                                  {KeySpace, "Space"},
                                                                                                  {KeyT, "T"},
                                                                                                  {KeyTab, "Tab"},
                                                                                                  {KeyU, "U"},
                                                                                                  {KeyUp, "Up"},
                                                                                                  {KeyV, "V"},
                                                                                                  {KeyVolumeDown, "VolumeDown"},
                                                                                                  {KeyVolumeUp, "VolumeUp"},
                                                                                                  {KeyW, "W"},
                                                                                                  {KeyX, "X"},
                                                                                                  {KeyY, "Y"},
                                                                                                  {KeyZ, "Z"}}};

        [[nodiscard]] inline std::optional<std::string_view> keyname(std::uint64_t key) noexcept
        {
            auto it = std::lower_bound(KEY_NAME.begin(), KEY_NAME.end(), key,
                                       [](const auto &p, std::uint64_t k)
                                       { return p.first < k; });
            if (it != KEY_NAME.end() && it->first == key)
            {
                return it->second;
            }
            return std::nullopt;
        }

        inline const std::unordered_map<std::uint8_t, std::uint8_t> CHAR_TO_KEY = {
            {static_cast<std::uint8_t>(0x09), static_cast<std::uint8_t>(KeyTab)},
            {static_cast<std::uint8_t>(0x0A), static_cast<std::uint8_t>(KeyEnter)},
            {static_cast<std::uint8_t>(0x0D), static_cast<std::uint8_t>(KeyEnter)},
            {static_cast<std::uint8_t>(0x20), static_cast<std::uint8_t>(KeySpace)},
            {static_cast<std::uint8_t>(0x21), static_cast<std::uint8_t>(Key1)},
            {static_cast<std::uint8_t>(0x22), static_cast<std::uint8_t>(KeyApostrophe)},
            {static_cast<std::uint8_t>(0x23), static_cast<std::uint8_t>(Key3)},
            {static_cast<std::uint8_t>(0x24), static_cast<std::uint8_t>(Key4)},
            {static_cast<std::uint8_t>(0x25), static_cast<std::uint8_t>(Key5)},
            {static_cast<std::uint8_t>(0x26), static_cast<std::uint8_t>(Key7)},
            {static_cast<std::uint8_t>(0x27), static_cast<std::uint8_t>(KeyApostrophe)},
            {static_cast<std::uint8_t>(0x28), static_cast<std::uint8_t>(Key9)},
            {static_cast<std::uint8_t>(0x29), static_cast<std::uint8_t>(Key0)},
            {static_cast<std::uint8_t>(0x2A), static_cast<std::uint8_t>(Key8)},
            {static_cast<std::uint8_t>(0x2B), static_cast<std::uint8_t>(KeyEqual)},
            {static_cast<std::uint8_t>(0x2C), static_cast<std::uint8_t>(KeyComma)},
            {static_cast<std::uint8_t>(0x2D), static_cast<std::uint8_t>(KeyMinus)},
            {static_cast<std::uint8_t>(0x2E), static_cast<std::uint8_t>(KeyPeriod)},
            {static_cast<std::uint8_t>(0x2F), static_cast<std::uint8_t>(KeySlash)},
            {static_cast<std::uint8_t>(0x30), static_cast<std::uint8_t>(Key0)},
            {static_cast<std::uint8_t>(0x31), static_cast<std::uint8_t>(Key1)},
            {static_cast<std::uint8_t>(0x32), static_cast<std::uint8_t>(Key2)},
            {static_cast<std::uint8_t>(0x33), static_cast<std::uint8_t>(Key3)},
            {static_cast<std::uint8_t>(0x34), static_cast<std::uint8_t>(Key4)},
            {static_cast<std::uint8_t>(0x35), static_cast<std::uint8_t>(Key5)},
            {static_cast<std::uint8_t>(0x36), static_cast<std::uint8_t>(Key6)},
            {static_cast<std::uint8_t>(0x37), static_cast<std::uint8_t>(Key7)},
            {static_cast<std::uint8_t>(0x38), static_cast<std::uint8_t>(Key8)},
            {static_cast<std::uint8_t>(0x39), static_cast<std::uint8_t>(Key9)},
            {static_cast<std::uint8_t>(0x3A), static_cast<std::uint8_t>(KeySemicolon)},
            {static_cast<std::uint8_t>(0x3B), static_cast<std::uint8_t>(KeySemicolon)},
            {static_cast<std::uint8_t>(0x3C), static_cast<std::uint8_t>(KeyComma)},
            {static_cast<std::uint8_t>(0x3D), static_cast<std::uint8_t>(KeyEqual)},
            {static_cast<std::uint8_t>(0x3E), static_cast<std::uint8_t>(KeyPeriod)},
            {static_cast<std::uint8_t>(0x3F), static_cast<std::uint8_t>(KeySlash)},
            {static_cast<std::uint8_t>(0x40), static_cast<std::uint8_t>(Key2)},
            {static_cast<std::uint8_t>(0x41), static_cast<std::uint8_t>(KeyA)},
            {static_cast<std::uint8_t>(0x42), static_cast<std::uint8_t>(KeyB)},
            {static_cast<std::uint8_t>(0x43), static_cast<std::uint8_t>(KeyC)},
            {static_cast<std::uint8_t>(0x44), static_cast<std::uint8_t>(KeyD)},
            {static_cast<std::uint8_t>(0x45), static_cast<std::uint8_t>(KeyE)},
            {static_cast<std::uint8_t>(0x46), static_cast<std::uint8_t>(KeyF)},
            {static_cast<std::uint8_t>(0x47), static_cast<std::uint8_t>(KeyG)},
            {static_cast<std::uint8_t>(0x48), static_cast<std::uint8_t>(KeyH)},
            {static_cast<std::uint8_t>(0x49), static_cast<std::uint8_t>(KeyI)},
            {static_cast<std::uint8_t>(0x4A), static_cast<std::uint8_t>(KeyJ)},
            {static_cast<std::uint8_t>(0x4B), static_cast<std::uint8_t>(KeyK)},
            {static_cast<std::uint8_t>(0x4C), static_cast<std::uint8_t>(KeyL)},
            {static_cast<std::uint8_t>(0x4D), static_cast<std::uint8_t>(KeyM)},
            {static_cast<std::uint8_t>(0x4E), static_cast<std::uint8_t>(KeyN)},
            {static_cast<std::uint8_t>(0x4F), static_cast<std::uint8_t>(KeyO)},
            {static_cast<std::uint8_t>(0x50), static_cast<std::uint8_t>(KeyP)},
            {static_cast<std::uint8_t>(0x51), static_cast<std::uint8_t>(KeyQ)},
            {static_cast<std::uint8_t>(0x52), static_cast<std::uint8_t>(KeyR)},
            {static_cast<std::uint8_t>(0x53), static_cast<std::uint8_t>(KeyS)},
            {static_cast<std::uint8_t>(0x54), static_cast<std::uint8_t>(KeyT)},
            {static_cast<std::uint8_t>(0x55), static_cast<std::uint8_t>(KeyU)},
            {static_cast<std::uint8_t>(0x56), static_cast<std::uint8_t>(KeyV)},
            {static_cast<std::uint8_t>(0x57), static_cast<std::uint8_t>(KeyW)},
            {static_cast<std::uint8_t>(0x58), static_cast<std::uint8_t>(KeyX)},
            {static_cast<std::uint8_t>(0x59), static_cast<std::uint8_t>(KeyY)},
            {static_cast<std::uint8_t>(0x5A), static_cast<std::uint8_t>(KeyZ)},
            {static_cast<std::uint8_t>(0x5B), static_cast<std::uint8_t>(KeyLeftBrace)},
            {static_cast<std::uint8_t>(0x5C), static_cast<std::uint8_t>(KeyBackslash)},
            {static_cast<std::uint8_t>(0x5D), static_cast<std::uint8_t>(KeyRightBrace)},
            {static_cast<std::uint8_t>(0x5E), static_cast<std::uint8_t>(Key6)},
            {static_cast<std::uint8_t>(0x5F), static_cast<std::uint8_t>(KeyMinus)},
            {static_cast<std::uint8_t>(0x60), static_cast<std::uint8_t>(KeyGrave)},
            {static_cast<std::uint8_t>(0x61), static_cast<std::uint8_t>(KeyA)},
            {static_cast<std::uint8_t>(0x62), static_cast<std::uint8_t>(KeyB)},
            {static_cast<std::uint8_t>(0x63), static_cast<std::uint8_t>(KeyC)},
            {static_cast<std::uint8_t>(0x64), static_cast<std::uint8_t>(KeyD)},
            {static_cast<std::uint8_t>(0x65), static_cast<std::uint8_t>(KeyE)},
            {static_cast<std::uint8_t>(0x66), static_cast<std::uint8_t>(KeyF)},
            {static_cast<std::uint8_t>(0x67), static_cast<std::uint8_t>(KeyG)},
            {static_cast<std::uint8_t>(0x68), static_cast<std::uint8_t>(KeyH)},
            {static_cast<std::uint8_t>(0x69), static_cast<std::uint8_t>(KeyI)},
            {static_cast<std::uint8_t>(0x6A), static_cast<std::uint8_t>(KeyJ)},
            {static_cast<std::uint8_t>(0x6B), static_cast<std::uint8_t>(KeyK)},
            {static_cast<std::uint8_t>(0x6C), static_cast<std::uint8_t>(KeyL)},
            {static_cast<std::uint8_t>(0x6D), static_cast<std::uint8_t>(KeyM)},
            {static_cast<std::uint8_t>(0x6E), static_cast<std::uint8_t>(KeyN)},
            {static_cast<std::uint8_t>(0x6F), static_cast<std::uint8_t>(KeyO)},
            {static_cast<std::uint8_t>(0x70), static_cast<std::uint8_t>(KeyP)},
            {static_cast<std::uint8_t>(0x71), static_cast<std::uint8_t>(KeyQ)},
            {static_cast<std::uint8_t>(0x72), static_cast<std::uint8_t>(KeyR)},
            {static_cast<std::uint8_t>(0x73), static_cast<std::uint8_t>(KeyS)},
            {static_cast<std::uint8_t>(0x74), static_cast<std::uint8_t>(KeyT)},
            {static_cast<std::uint8_t>(0x75), static_cast<std::uint8_t>(KeyU)},
            {static_cast<std::uint8_t>(0x76), static_cast<std::uint8_t>(KeyV)},
            {static_cast<std::uint8_t>(0x77), static_cast<std::uint8_t>(KeyW)},
            {static_cast<std::uint8_t>(0x78), static_cast<std::uint8_t>(KeyX)},
            {static_cast<std::uint8_t>(0x79), static_cast<std::uint8_t>(KeyY)},
            {static_cast<std::uint8_t>(0x7A), static_cast<std::uint8_t>(KeyZ)},
            {static_cast<std::uint8_t>(0x7B), static_cast<std::uint8_t>(KeyLeftBrace)},
            {static_cast<std::uint8_t>(0x7C), static_cast<std::uint8_t>(KeyBackslash)},
            {static_cast<std::uint8_t>(0x7D), static_cast<std::uint8_t>(KeyRightBrace)},
            {static_cast<std::uint8_t>(0x7E), static_cast<std::uint8_t>(KeyGrave)},
        };

        inline const std::unordered_set<std::uint8_t> SHIFT_CHARS = {
            static_cast<std::uint8_t>(0x21),
            static_cast<std::uint8_t>(0x22),
            static_cast<std::uint8_t>(0x23),
            static_cast<std::uint8_t>(0x24),
            static_cast<std::uint8_t>(0x25),
            static_cast<std::uint8_t>(0x26),
            static_cast<std::uint8_t>(0x28),
            static_cast<std::uint8_t>(0x29),
            static_cast<std::uint8_t>(0x2A),
            static_cast<std::uint8_t>(0x2B),
            static_cast<std::uint8_t>(0x3A),
            static_cast<std::uint8_t>(0x3C),
            static_cast<std::uint8_t>(0x3E),
            static_cast<std::uint8_t>(0x3F),
            static_cast<std::uint8_t>(0x40),
            static_cast<std::uint8_t>(0x41),
            static_cast<std::uint8_t>(0x42),
            static_cast<std::uint8_t>(0x43),
            static_cast<std::uint8_t>(0x44),
            static_cast<std::uint8_t>(0x45),
            static_cast<std::uint8_t>(0x46),
            static_cast<std::uint8_t>(0x47),
            static_cast<std::uint8_t>(0x48),
            static_cast<std::uint8_t>(0x49),
            static_cast<std::uint8_t>(0x4A),
            static_cast<std::uint8_t>(0x4B),
            static_cast<std::uint8_t>(0x4C),
            static_cast<std::uint8_t>(0x4D),
            static_cast<std::uint8_t>(0x4E),
            static_cast<std::uint8_t>(0x4F),
            static_cast<std::uint8_t>(0x50),
            static_cast<std::uint8_t>(0x51),
            static_cast<std::uint8_t>(0x52),
            static_cast<std::uint8_t>(0x53),
            static_cast<std::uint8_t>(0x54),
            static_cast<std::uint8_t>(0x55),
            static_cast<std::uint8_t>(0x56),
            static_cast<std::uint8_t>(0x57),
            static_cast<std::uint8_t>(0x58),
            static_cast<std::uint8_t>(0x59),
            static_cast<std::uint8_t>(0x5A),
            static_cast<std::uint8_t>(0x5E),
            static_cast<std::uint8_t>(0x5F),
            static_cast<std::uint8_t>(0x7B),
            static_cast<std::uint8_t>(0x7C),
            static_cast<std::uint8_t>(0x7D),
            static_cast<std::uint8_t>(0x7E),
        };

        // ============================================================================
        // Input: Client -> Device
        // ============================================================================

        struct Input
        {
            std::uint8_t modifiers = 0;
            std::vector<std::uint8_t> keys;

            [[nodiscard]] std::vector<std::uint8_t> to_bytes() const
            {
                std::vector<std::uint8_t> buf;
                buf.push_back(modifiers);
                buf.push_back(static_cast<std::uint8_t>(keys.size()));
                for (const auto &v : keys)
                {
                    buf.push_back(static_cast<std::uint8_t>(v));
                }
                return buf;
            }
        };

        // ============================================================================
        // Output: Device -> Client
        // ============================================================================

        struct Output
        {
            std::uint8_t leds = 0;

            static Result<Output> from_bytes(const std::uint8_t *data, std::size_t len)
            {
                Output result;
                std::size_t offset = 0;
                if (offset >= len)
                    return Error("buffer too short");
                result.leds = data[offset++];
                (void)offset; // suppress unused warning
                return result;
            }
        };

    } // namespace keyboard
} // namespace viiper
