using System;
using System.Collections.Generic;

namespace Viiper.Client.Devices.Keyboard;


/// <summary>
/// Size in bytes of Keyboard output (server-to-client) messages.
/// Use this constant to allocate buffers for reading device output.
/// </summary>
public static class Keyboard
{

    public const int OutputSize = 1;


}



/// <summary>
/// Key constants for Keyboard device.
/// </summary>
[Flags]
public enum Key : byte
{
    A = 0x4,
    B = 0x5,
    C = 0x6,
    D = 0x7,
    E = 0x8,
    F = 0x9,
    G = 0xA,
    H = 0xB,
    I = 0xC,
    J = 0xD,
    K = 0xE,
    L = 0xF,
    M = 0x10,
    N = 0x11,
    O = 0x12,
    P = 0x13,
    Q = 0x14,
    R = 0x15,
    S = 0x16,
    T = 0x17,
    U = 0x18,
    V = 0x19,
    W = 0x1A,
    X = 0x1B,
    Y = 0x1C,
    Z = 0x1D,
    Num1 = 0x1E,
    Num2 = 0x1F,
    Num3 = 0x20,
    Num4 = 0x21,
    Num5 = 0x22,
    Num6 = 0x23,
    Num7 = 0x24,
    Num8 = 0x25,
    Num9 = 0x26,
    Num0 = 0x27,
    Enter = 0x28,
    Escape = 0x29,
    Backspace = 0x2A,
    Tab = 0x2B,
    Space = 0x2C,
    Minus = 0x2D,
    Equal = 0x2E,
    LeftBrace = 0x2F,
    RightBrace = 0x30,
    Backslash = 0x31,
    NonUSHash = 0x32,
    Semicolon = 0x33,
    Apostrophe = 0x34,
    Grave = 0x35,
    Comma = 0x36,
    Period = 0x37,
    Slash = 0x38,
    CapsLock = 0x39,
    F1 = 0x3A,
    F2 = 0x3B,
    F3 = 0x3C,
    F4 = 0x3D,
    F5 = 0x3E,
    F6 = 0x3F,
    F7 = 0x40,
    F8 = 0x41,
    F9 = 0x42,
    F10 = 0x43,
    F11 = 0x44,
    F12 = 0x45,
    PrintScreen = 0x46,
    ScrollLock = 0x47,
    Pause = 0x48,
    Insert = 0x49,
    Home = 0x4A,
    PageUp = 0x4B,
    Delete = 0x4C,
    End = 0x4D,
    PageDown = 0x4E,
    Right = 0x4F,
    Left = 0x50,
    Down = 0x51,
    Up = 0x52,
    NumLock = 0x53,
    KpSlash = 0x54,
    KpAsterisk = 0x55,
    KpMinus = 0x56,
    KpPlus = 0x57,
    KpEnter = 0x58,
    Kp1 = 0x59,
    Kp2 = 0x5A,
    Kp3 = 0x5B,
    Kp4 = 0x5C,
    Kp5 = 0x5D,
    Kp6 = 0x5E,
    Kp7 = 0x5F,
    Kp8 = 0x60,
    Kp9 = 0x61,
    Kp0 = 0x62,
    KpDot = 0x63,
    NonUSBackslash = 0x64,
    Application = 0x65,
    Power = 0x66,
    KpEqual = 0x67,
    F13 = 0x68,
    F14 = 0x69,
    F15 = 0x6A,
    F16 = 0x6B,
    F17 = 0x6C,
    F18 = 0x6D,
    F19 = 0x6E,
    F20 = 0x6F,
    F21 = 0x70,
    F22 = 0x71,
    F23 = 0x72,
    F24 = 0x73,
    Execute = 0x74,
    Help = 0x75,
    Menu = 0x76,
    Select = 0x77,
    Stop = 0x78,
    Again = 0x79,
    Undo = 0x7A,
    Cut = 0x7B,
    Copy = 0x7C,
    Paste = 0x7D,
    Find = 0x7E,
    Mute = 0x7F,
    VolumeUp = 0x80,
    VolumeDown = 0x81,
    MediaPlayPause = 0xE8,
    MediaStop = 0xE9,
    MediaNext = 0xEB,
    MediaPrevious = 0xEC,
}


/// <summary>
/// LED constants for Keyboard device.
/// </summary>
[Flags]
public enum LED : byte
{
    NumLock = 0x1,
    CapsLock = 0x2,
    ScrollLock = 0x4,
    Compose = 0x8,
    Kana = 0x10,
}


/// <summary>
/// Mod constants for Keyboard device.
/// </summary>
[Flags]
public enum Mod : byte
{
    LeftCtrl = 0x1,
    LeftShift = 0x2,
    LeftAlt = 0x4,
    LeftGUI = 0x8,
    RightCtrl = 0x10,
    RightShift = 0x20,
    RightAlt = 0x40,
    RightGUI = 0x80,
}



/// <summary>
/// KeyName lookup map for Keyboard device.
/// </summary>
public static class KeyName
{
    private static readonly Dictionary<byte, string> _map = new()
    {
        { (byte)Key.Num0, "0" },
        { (byte)Key.Num1, "1" },
        { (byte)Key.Num2, "2" },
        { (byte)Key.Num3, "3" },
        { (byte)Key.Num4, "4" },
        { (byte)Key.Num5, "5" },
        { (byte)Key.Num6, "6" },
        { (byte)Key.Num7, "7" },
        { (byte)Key.Num8, "8" },
        { (byte)Key.Num9, "9" },
        { (byte)Key.A, "A" },
        { (byte)Key.Apostrophe, "Apostrophe" },
        { (byte)Key.Application, "Application" },
        { (byte)Key.B, "B" },
        { (byte)Key.Backslash, "Backslash" },
        { (byte)Key.Backspace, "Backspace" },
        { (byte)Key.C, "C" },
        { (byte)Key.CapsLock, "CapsLock" },
        { (byte)Key.Comma, "Comma" },
        { (byte)Key.D, "D" },
        { (byte)Key.Delete, "Delete" },
        { (byte)Key.Down, "Down" },
        { (byte)Key.E, "E" },
        { (byte)Key.End, "End" },
        { (byte)Key.Enter, "Enter" },
        { (byte)Key.Equal, "Equal" },
        { (byte)Key.Escape, "Escape" },
        { (byte)Key.F, "F" },
        { (byte)Key.F1, "F1" },
        { (byte)Key.F10, "F10" },
        { (byte)Key.F11, "F11" },
        { (byte)Key.F12, "F12" },
        { (byte)Key.F13, "F13" },
        { (byte)Key.F14, "F14" },
        { (byte)Key.F15, "F15" },
        { (byte)Key.F16, "F16" },
        { (byte)Key.F17, "F17" },
        { (byte)Key.F18, "F18" },
        { (byte)Key.F19, "F19" },
        { (byte)Key.F2, "F2" },
        { (byte)Key.F20, "F20" },
        { (byte)Key.F21, "F21" },
        { (byte)Key.F22, "F22" },
        { (byte)Key.F23, "F23" },
        { (byte)Key.F24, "F24" },
        { (byte)Key.F3, "F3" },
        { (byte)Key.F4, "F4" },
        { (byte)Key.F5, "F5" },
        { (byte)Key.F6, "F6" },
        { (byte)Key.F7, "F7" },
        { (byte)Key.F8, "F8" },
        { (byte)Key.F9, "F9" },
        { (byte)Key.G, "G" },
        { (byte)Key.Grave, "Grave" },
        { (byte)Key.H, "H" },
        { (byte)Key.Home, "Home" },
        { (byte)Key.I, "I" },
        { (byte)Key.Insert, "Insert" },
        { (byte)Key.J, "J" },
        { (byte)Key.K, "K" },
        { (byte)Key.Kp0, "Kp0" },
        { (byte)Key.Kp1, "Kp1" },
        { (byte)Key.Kp2, "Kp2" },
        { (byte)Key.Kp3, "Kp3" },
        { (byte)Key.Kp4, "Kp4" },
        { (byte)Key.Kp5, "Kp5" },
        { (byte)Key.Kp6, "Kp6" },
        { (byte)Key.Kp7, "Kp7" },
        { (byte)Key.Kp8, "Kp8" },
        { (byte)Key.Kp9, "Kp9" },
        { (byte)Key.KpAsterisk, "Kp*" },
        { (byte)Key.KpDot, "Kp." },
        { (byte)Key.KpEnter, "KpEnter" },
        { (byte)Key.KpMinus, "Kp-" },
        { (byte)Key.KpPlus, "Kp+" },
        { (byte)Key.KpSlash, "Kp/" },
        { (byte)Key.L, "L" },
        { (byte)Key.Left, "Left" },
        { (byte)Key.LeftBrace, "LeftBrace" },
        { (byte)Key.M, "M" },
        { (byte)Key.MediaNext, "MediaNext" },
        { (byte)Key.MediaPlayPause, "MediaPlayPause" },
        { (byte)Key.MediaPrevious, "MediaPrevious" },
        { (byte)Key.MediaStop, "MediaStop" },
        { (byte)Key.Minus, "Minus" },
        { (byte)Key.Mute, "Mute" },
        { (byte)Key.N, "N" },
        { (byte)Key.NumLock, "NumLock" },
        { (byte)Key.O, "O" },
        { (byte)Key.P, "P" },
        { (byte)Key.PageDown, "PageDown" },
        { (byte)Key.PageUp, "PageUp" },
        { (byte)Key.Pause, "Pause" },
        { (byte)Key.Period, "Period" },
        { (byte)Key.PrintScreen, "PrintScreen" },
        { (byte)Key.Q, "Q" },
        { (byte)Key.R, "R" },
        { (byte)Key.Right, "Right" },
        { (byte)Key.RightBrace, "RightBrace" },
        { (byte)Key.S, "S" },
        { (byte)Key.ScrollLock, "ScrollLock" },
        { (byte)Key.Semicolon, "Semicolon" },
        { (byte)Key.Slash, "Slash" },
        { (byte)Key.Space, "Space" },
        { (byte)Key.T, "T" },
        { (byte)Key.Tab, "Tab" },
        { (byte)Key.U, "U" },
        { (byte)Key.Up, "Up" },
        { (byte)Key.V, "V" },
        { (byte)Key.VolumeDown, "VolumeDown" },
        { (byte)Key.VolumeUp, "VolumeUp" },
        { (byte)Key.W, "W" },
        { (byte)Key.X, "X" },
        { (byte)Key.Y, "Y" },
        { (byte)Key.Z, "Z" },
    };

    /// <summary>
    /// Try to get the value for the given key.
    /// </summary>
    public static bool TryGetValue(byte key, out string value)
    {
        return _map.TryGetValue(key, out value);
    }

    /// <summary>
    /// Get the value for the given key, or return the default value if not found.
    /// </summary>
	public static string? GetValueOrDefault(byte key, string? defaultValue = default)
    {
        return _map.TryGetValue(key, out var value) ? value : defaultValue;
    }

    /// <summary>
    /// Check if the map contains the given key.
    /// </summary>
    public static bool ContainsKey(byte key)
    {
        return _map.ContainsKey(key);
    }
}


/// <summary>
/// CharToKey lookup map for Keyboard device.
/// </summary>
public static class CharToKey
{
    private static readonly Dictionary<byte, Key> _map = new()
    {
        { (byte)0x09, Key.Tab },
        { (byte)0x0A, Key.Enter },
        { (byte)0x0D, Key.Enter },
        { (byte)' ', Key.Space },
        { (byte)'!', Key.Num1 },
        { (byte)'"', Key.Apostrophe },
        { (byte)'#', Key.Num3 },
        { (byte)'$', Key.Num4 },
        { (byte)'%', Key.Num5 },
        { (byte)'&', Key.Num7 },
        { (byte)'\'', Key.Apostrophe },
        { (byte)'(', Key.Num9 },
        { (byte)')', Key.Num0 },
        { (byte)'*', Key.Num8 },
        { (byte)'+', Key.Equal },
        { (byte)',', Key.Comma },
        { (byte)'-', Key.Minus },
        { (byte)'.', Key.Period },
        { (byte)'/', Key.Slash },
        { (byte)'0', Key.Num0 },
        { (byte)'1', Key.Num1 },
        { (byte)'2', Key.Num2 },
        { (byte)'3', Key.Num3 },
        { (byte)'4', Key.Num4 },
        { (byte)'5', Key.Num5 },
        { (byte)'6', Key.Num6 },
        { (byte)'7', Key.Num7 },
        { (byte)'8', Key.Num8 },
        { (byte)'9', Key.Num9 },
        { (byte)':', Key.Semicolon },
        { (byte)';', Key.Semicolon },
        { (byte)'<', Key.Comma },
        { (byte)'=', Key.Equal },
        { (byte)'>', Key.Period },
        { (byte)'?', Key.Slash },
        { (byte)'@', Key.Num2 },
        { (byte)'A', Key.A },
        { (byte)'B', Key.B },
        { (byte)'C', Key.C },
        { (byte)'D', Key.D },
        { (byte)'E', Key.E },
        { (byte)'F', Key.F },
        { (byte)'G', Key.G },
        { (byte)'H', Key.H },
        { (byte)'I', Key.I },
        { (byte)'J', Key.J },
        { (byte)'K', Key.K },
        { (byte)'L', Key.L },
        { (byte)'M', Key.M },
        { (byte)'N', Key.N },
        { (byte)'O', Key.O },
        { (byte)'P', Key.P },
        { (byte)'Q', Key.Q },
        { (byte)'R', Key.R },
        { (byte)'S', Key.S },
        { (byte)'T', Key.T },
        { (byte)'U', Key.U },
        { (byte)'V', Key.V },
        { (byte)'W', Key.W },
        { (byte)'X', Key.X },
        { (byte)'Y', Key.Y },
        { (byte)'Z', Key.Z },
        { (byte)'[', Key.LeftBrace },
        { (byte)'\\', Key.Backslash },
        { (byte)']', Key.RightBrace },
        { (byte)'^', Key.Num6 },
        { (byte)'_', Key.Minus },
        { (byte)'`', Key.Grave },
        { (byte)'a', Key.A },
        { (byte)'b', Key.B },
        { (byte)'c', Key.C },
        { (byte)'d', Key.D },
        { (byte)'e', Key.E },
        { (byte)'f', Key.F },
        { (byte)'g', Key.G },
        { (byte)'h', Key.H },
        { (byte)'i', Key.I },
        { (byte)'j', Key.J },
        { (byte)'k', Key.K },
        { (byte)'l', Key.L },
        { (byte)'m', Key.M },
        { (byte)'n', Key.N },
        { (byte)'o', Key.O },
        { (byte)'p', Key.P },
        { (byte)'q', Key.Q },
        { (byte)'r', Key.R },
        { (byte)'s', Key.S },
        { (byte)'t', Key.T },
        { (byte)'u', Key.U },
        { (byte)'v', Key.V },
        { (byte)'w', Key.W },
        { (byte)'x', Key.X },
        { (byte)'y', Key.Y },
        { (byte)'z', Key.Z },
        { (byte)'{', Key.LeftBrace },
        { (byte)'|', Key.Backslash },
        { (byte)'}', Key.RightBrace },
        { (byte)'~', Key.Grave },
    };

    /// <summary>
    /// Try to get the value for the given key.
    /// </summary>
    public static bool TryGetValue(byte key, out Key value)
    {
        return _map.TryGetValue(key, out value);
    }

    /// <summary>
    /// Get the value for the given key, or return the default value if not found.
    /// </summary>
	public static Key GetValueOrDefault(byte key, Key defaultValue = default)
    {
        return _map.TryGetValue(key, out var value) ? value : defaultValue;
    }

    /// <summary>
    /// Check if the map contains the given key.
    /// </summary>
    public static bool ContainsKey(byte key)
    {
        return _map.ContainsKey(key);
    }
}


/// <summary>
/// ShiftChars lookup map for Keyboard device.
/// </summary>
public static class ShiftChars
{
    private static readonly Dictionary<byte, bool> _map = new()
    {
        { (byte)'!', true },
        { (byte)'"', true },
        { (byte)'#', true },
        { (byte)'$', true },
        { (byte)'%', true },
        { (byte)'&', true },
        { (byte)'(', true },
        { (byte)')', true },
        { (byte)'*', true },
        { (byte)'+', true },
        { (byte)':', true },
        { (byte)'<', true },
        { (byte)'>', true },
        { (byte)'?', true },
        { (byte)'@', true },
        { (byte)'A', true },
        { (byte)'B', true },
        { (byte)'C', true },
        { (byte)'D', true },
        { (byte)'E', true },
        { (byte)'F', true },
        { (byte)'G', true },
        { (byte)'H', true },
        { (byte)'I', true },
        { (byte)'J', true },
        { (byte)'K', true },
        { (byte)'L', true },
        { (byte)'M', true },
        { (byte)'N', true },
        { (byte)'O', true },
        { (byte)'P', true },
        { (byte)'Q', true },
        { (byte)'R', true },
        { (byte)'S', true },
        { (byte)'T', true },
        { (byte)'U', true },
        { (byte)'V', true },
        { (byte)'W', true },
        { (byte)'X', true },
        { (byte)'Y', true },
        { (byte)'Z', true },
        { (byte)'^', true },
        { (byte)'_', true },
        { (byte)'{', true },
        { (byte)'|', true },
        { (byte)'}', true },
        { (byte)'~', true },
    };

    /// <summary>
    /// Try to get the value for the given key.
    /// </summary>
    public static bool TryGetValue(byte key, out bool value)
    {
        return _map.TryGetValue(key, out value);
    }

    /// <summary>
    /// Get the value for the given key, or return the default value if not found.
    /// </summary>
	public static bool GetValueOrDefault(byte key, bool defaultValue = default)
    {
        return _map.TryGetValue(key, out var value) ? value : defaultValue;
    }

    /// <summary>
    /// Check if the map contains the given key.
    /// </summary>
    public static bool ContainsKey(byte key)
    {
        return _map.ContainsKey(key);
    }
}


