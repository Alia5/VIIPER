package sdl

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../_testing/e2e/deps/SDL/include

#include <stdlib.h>

#include <SDL3/SDL_hints.h>
*/
import "C"
import "unsafe"

// HintPriority is an enumeration of hint priorities.
//
// \since This enum is available since SDL 3.2.0.
type HintPriority int

const (
	HintPriorityDefault  HintPriority = C.SDL_HINT_DEFAULT
	HintPriorityNormal   HintPriority = C.SDL_HINT_NORMAL
	HintPriorityOverride HintPriority = C.SDL_HINT_OVERRIDE
)

// Hint is the name of an SDL configuration hint.
//
// The convention for naming hints is SDL_HINT_X, where "SDL_X" is the
// environment variable that can be used to override the default.
//
// In general these hints are just that - they may or may not be supported or
// applicable on any given platform, but they provide a way for an application
// or user to give the library a hint as to how they would like the library to
// work.
type Hint string

// Hints mirrors SDL_HINT_* from SDL_hints.h using Go-style names.
var (
	HintAllowAltTabWhileGrabbed            Hint = Hint(C.SDL_HINT_ALLOW_ALT_TAB_WHILE_GRABBED)
	HintAndroidAllowRecreateActivity       Hint = Hint(C.SDL_HINT_ANDROID_ALLOW_RECREATE_ACTIVITY)
	HintAndroidBlockOnPause                Hint = Hint(C.SDL_HINT_ANDROID_BLOCK_ON_PAUSE)
	HintAndroidLowLatencyAudio             Hint = Hint(C.SDL_HINT_ANDROID_LOW_LATENCY_AUDIO)
	HintAndroidTrapBackButton              Hint = Hint(C.SDL_HINT_ANDROID_TRAP_BACK_BUTTON)
	HintAndroidAllowPersistentFolderAccess Hint = Hint(C.SDL_HINT_ANDROID_ALLOW_PERSISTENT_FOLDER_ACCESS)
	HintAppID                              Hint = Hint(C.SDL_HINT_APP_ID)
	HintAppName                            Hint = Hint(C.SDL_HINT_APP_NAME)
	HintAppleTvControllerUIEvents          Hint = Hint(C.SDL_HINT_APPLE_TV_CONTROLLER_UI_EVENTS)
	HintAppleTvRemoteAllowRotation         Hint = Hint(C.SDL_HINT_APPLE_TV_REMOTE_ALLOW_ROTATION)
	HintAudioAlsaDefaultDevice             Hint = Hint(C.SDL_HINT_AUDIO_ALSA_DEFAULT_DEVICE)
	HintAudioAlsaDefaultPlaybackDevice     Hint = Hint(C.SDL_HINT_AUDIO_ALSA_DEFAULT_PLAYBACK_DEVICE)
	HintAudioAlsaDefaultRecordingDevice    Hint = Hint(C.SDL_HINT_AUDIO_ALSA_DEFAULT_RECORDING_DEVICE)
	HintAudioCategory                      Hint = Hint(C.SDL_HINT_AUDIO_CATEGORY)
	HintAudioChannels                      Hint = Hint(C.SDL_HINT_AUDIO_CHANNELS)
	HintAudioDeviceAppIconName             Hint = Hint(C.SDL_HINT_AUDIO_DEVICE_APP_ICON_NAME)
	HintAudioDeviceSampleFrames            Hint = Hint(C.SDL_HINT_AUDIO_DEVICE_SAMPLE_FRAMES)
	HintAudioDeviceStreamName              Hint = Hint(C.SDL_HINT_AUDIO_DEVICE_STREAM_NAME)
	HintAudioDeviceStreamRole              Hint = Hint(C.SDL_HINT_AUDIO_DEVICE_STREAM_ROLE)
	HintAudioDeviceRawStream               Hint = Hint(C.SDL_HINT_AUDIO_DEVICE_RAW_STREAM)
	HintAudioDiskInputFile                 Hint = Hint(C.SDL_HINT_AUDIO_DISK_INPUT_FILE)
	HintAudioDiskOutputFile                Hint = Hint(C.SDL_HINT_AUDIO_DISK_OUTPUT_FILE)
	HintAudioDiskTimescale                 Hint = Hint(C.SDL_HINT_AUDIO_DISK_TIMESCALE)
	HintAudioDriver                        Hint = Hint(C.SDL_HINT_AUDIO_DRIVER)
	HintAudioDummyTimescale                Hint = Hint(C.SDL_HINT_AUDIO_DUMMY_TIMESCALE)
	HintAudioFormat                        Hint = Hint(C.SDL_HINT_AUDIO_FORMAT)
	HintAudioFrequency                     Hint = Hint(C.SDL_HINT_AUDIO_FREQUENCY)
	HintAudioIncludeMonitors               Hint = Hint(C.SDL_HINT_AUDIO_INCLUDE_MONITORS)
	HintAutoUpdateJoysticks                Hint = Hint(C.SDL_HINT_AUTO_UPDATE_JOYSTICKS)
	HintAutoUpdateSensors                  Hint = Hint(C.SDL_HINT_AUTO_UPDATE_SENSORS)
	HintBmPSaveLegacyFormat                Hint = Hint(C.SDL_HINT_BMP_SAVE_LEGACY_FORMAT)
	HintCameraDriver                       Hint = Hint(C.SDL_HINT_CAMERA_DRIVER)
	HintCPUFeatureMask                     Hint = Hint(C.SDL_HINT_CPU_FEATURE_MASK)
	HintJoystickDirectinput                Hint = Hint(C.SDL_HINT_JOYSTICK_DIRECTINPUT)
	HintFileDialogDriver                   Hint = Hint(C.SDL_HINT_FILE_DIALOG_DRIVER)
	HintDisplayUsableBounds                Hint = Hint(C.SDL_HINT_DISPLAY_USABLE_BOUNDS)
	HintDosAllowDirectFramebuffer          Hint = Hint(C.SDL_HINT_DOS_ALLOW_DIRECT_FRAMEBUFFER)
	HintInvalidParamChecks                 Hint = Hint(C.SDL_HINT_INVALID_PARAM_CHECKS)
	HintEmscriptenAsyncify                 Hint = Hint(C.SDL_HINT_EMSCRIPTEN_ASYNCIFY)
	HintEmscriptenCanvasSelector           Hint = Hint(C.SDL_HINT_EMSCRIPTEN_CANVAS_SELECTOR)
	HintEmscriptenKeyboardElement          Hint = Hint(C.SDL_HINT_EMSCRIPTEN_KEYBOARD_ELEMENT)
	HintEnableScreenKeyboard               Hint = Hint(C.SDL_HINT_ENABLE_SCREEN_KEYBOARD)
	HintEvdevDevices                       Hint = Hint(C.SDL_HINT_EVDEV_DEVICES)
	HintEventLogging                       Hint = Hint(C.SDL_HINT_EVENT_LOGGING)
	HintForceRaisewindow                   Hint = Hint(C.SDL_HINT_FORCE_RAISEWINDOW)
	HintFramebufferAcceleration            Hint = Hint(C.SDL_HINT_FRAMEBUFFER_ACCELERATION)
	HintGamecontrollerconfig               Hint = Hint(C.SDL_HINT_GAMECONTROLLERCONFIG)
	HintGamecontrollerconfigFile           Hint = Hint(C.SDL_HINT_GAMECONTROLLERCONFIG_FILE)
	HintGamecontrollertype                 Hint = Hint(C.SDL_HINT_GAMECONTROLLERTYPE)
	HintGamecontrollerIgnoreDevices        Hint = Hint(C.SDL_HINT_GAMECONTROLLER_IGNORE_DEVICES)
	HintGamecontrollerIgnoreDevicesExcept  Hint = Hint(C.SDL_HINT_GAMECONTROLLER_IGNORE_DEVICES_EXCEPT)
	HintGamecontrollerSensorFusion         Hint = Hint(C.SDL_HINT_GAMECONTROLLER_SENSOR_FUSION)
	HintGdkTextinputDefaultText            Hint = Hint(C.SDL_HINT_GDK_TEXTINPUT_DEFAULT_TEXT)
	HintGdkTextinputDescription            Hint = Hint(C.SDL_HINT_GDK_TEXTINPUT_DESCRIPTION)
	HintGdkTextinputMaxLength              Hint = Hint(C.SDL_HINT_GDK_TEXTINPUT_MAX_LENGTH)
	HintGdkTextinputScope                  Hint = Hint(C.SDL_HINT_GDK_TEXTINPUT_SCOPE)
	HintGdkTextinputTitle                  Hint = Hint(C.SDL_HINT_GDK_TEXTINPUT_TITLE)
	HintHIDAPILibusb                       Hint = Hint(C.SDL_HINT_HIDAPI_LIBUSB)
	HintHIDAPILibusbGamecube               Hint = Hint(C.SDL_HINT_HIDAPI_LIBUSB_GAMECUBE)
	HintHIDAPILibusbWhitelist              Hint = Hint(C.SDL_HINT_HIDAPI_LIBUSB_WHITELIST)
	HintHIDAPIUdev                         Hint = Hint(C.SDL_HINT_HIDAPI_UDEV)
	HintGpuDriver                          Hint = Hint(C.SDL_HINT_GPU_DRIVER)
	HintOpenxrLibrary                      Hint = Hint(C.SDL_HINT_OPENXR_LIBRARY)
	HintHIDAPIEnumerateOnlyControllers     Hint = Hint(C.SDL_HINT_HIDAPI_ENUMERATE_ONLY_CONTROLLERS)
	HintHIDAPIIgnoreDevices                Hint = Hint(C.SDL_HINT_HIDAPI_IGNORE_DEVICES)
	HintImeImplementedUI                   Hint = Hint(C.SDL_HINT_IME_IMPLEMENTED_UI)
	HintIosHideHomeIndicator               Hint = Hint(C.SDL_HINT_IOS_HIDE_HOME_INDICATOR)
	HintJoystickAllowBackgroundEvents      Hint = Hint(C.SDL_HINT_JOYSTICK_ALLOW_BACKGROUND_EVENTS)
	HintJoystickArcadestickDevices         Hint = Hint(C.SDL_HINT_JOYSTICK_ARCADESTICK_DEVICES)
	HintJoystickArcadestickDevicesExcluded Hint = Hint(C.SDL_HINT_JOYSTICK_ARCADESTICK_DEVICES_EXCLUDED)
	HintJoystickBlacklistDevices           Hint = Hint(C.SDL_HINT_JOYSTICK_BLACKLIST_DEVICES)
	HintJoystickBlacklistDevicesExcluded   Hint = Hint(C.SDL_HINT_JOYSTICK_BLACKLIST_DEVICES_EXCLUDED)
	HintJoystickDevice                     Hint = Hint(C.SDL_HINT_JOYSTICK_DEVICE)
	HintJoystickDrumDevices                Hint = Hint(C.SDL_HINT_JOYSTICK_DRUM_DEVICES)
	HintJoystickEnhancedReports            Hint = Hint(C.SDL_HINT_JOYSTICK_ENHANCED_REPORTS)
	HintJoystickFlightstickDevices         Hint = Hint(C.SDL_HINT_JOYSTICK_FLIGHTSTICK_DEVICES)
	HintJoystickFlightstickDevicesExcluded Hint = Hint(C.SDL_HINT_JOYSTICK_FLIGHTSTICK_DEVICES_EXCLUDED)
	HintJoystickGameinput                  Hint = Hint(C.SDL_HINT_JOYSTICK_GAMEINPUT)
	HintJoystickGameinputRaw               Hint = Hint(C.SDL_HINT_JOYSTICK_GAMEINPUT_RAW)
	HintJoystickGamecubeDevices            Hint = Hint(C.SDL_HINT_JOYSTICK_GAMECUBE_DEVICES)
	HintJoystickGamecubeDevicesExcluded    Hint = Hint(C.SDL_HINT_JOYSTICK_GAMECUBE_DEVICES_EXCLUDED)
	HintJoystickGuitarDevices              Hint = Hint(C.SDL_HINT_JOYSTICK_GUITAR_DEVICES)
	HintJoystickHIDAPI                     Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI)
	HintJoystickHIDAPICombineJoyCons       Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_COMBINE_JOY_CONS)
	HintJoystickHIDAPIGamecube             Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_GAMECUBE)
	HintJoystickHIDAPIGamecubeRumbleBrake  Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_GAMECUBE_RUMBLE_BRAKE)
	HintJoystickHIDAPIJoyCons              Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_JOY_CONS)
	HintJoystickHIDAPIJoyconHomeLed        Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_JOYCON_HOME_LED)
	HintJoystickHIDAPILuna                 Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_LUNA)
	HintJoystickHIDAPINintendoClassic      Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_NINTENDO_CLASSIC)
	HintJoystickHIDAPIPS3                  Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_PS3)
	HintJoystickHIDAPIPS3SixaxisDriver     Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_PS3_SIXAXIS_DRIVER)
	HintJoystickHIDAPIPS4                  Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_PS4)
	HintJoystickHIDAPIPS4ReportInterval    Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_PS4_REPORT_INTERVAL)
	HintJoystickHIDAPIPS5                  Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_PS5)
	HintJoystickHIDAPIPS5PlayerLed         Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_PS5_PLAYER_LED)
	HintJoystickHIDAPIShield               Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_SHIELD)
	HintJoystickHIDAPIStadia               Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_STADIA)
	HintJoystickHIDAPISteam                Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_STEAM)
	HintJoystickHIDAPISteamHomeLed         Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_STEAM_HOME_LED)
	HintJoystickHIDAPISteamdeck            Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_STEAMDECK)
	HintJoystickHIDAPISteamHori            Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_STEAM_HORI)
	HintJoystickHIDAPILg4ff                Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_LG4FF)
	HintJoystickHIDAPI8bitdo               Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_8BITDO)
	HintJoystickHIDAPISinput               Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_SINPUT)
	HintJoystickHIDAPIZuiki                Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_ZUIKI)
	HintJoystickHIDAPIFlydigi              Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_FLYDIGI)
	HintJoystickHIDAPIGamesir              Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_GAMESIR)
	HintJoystickHIDAPISwitch               Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_SWITCH)
	HintJoystickHIDAPISwitchHomeLed        Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_SWITCH_HOME_LED)
	HintJoystickHIDAPISwitchPlayerLed      Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_SWITCH_PLAYER_LED)
	HintJoystickHIDAPISwitch2              Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_SWITCH2)
	HintJoystickHIDAPIVerticalJoyCons      Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_VERTICAL_JOY_CONS)
	HintJoystickHIDAPIWii                  Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_WII)
	HintJoystickHIDAPIWiiPlayerLed         Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_WII_PLAYER_LED)
	HintJoystickHIDAPIXbox                 Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_XBOX)
	HintJoystickHIDAPIXbox360              Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_XBOX_360)
	HintJoystickHIDAPIXbox360PlayerLed     Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_XBOX_360_PLAYER_LED)
	HintJoystickHIDAPIXbox360Wireless      Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_XBOX_360_WIRELESS)
	HintJoystickHIDAPIXboxOne              Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_XBOX_ONE)
	HintJoystickHIDAPIXboxOneHomeLed       Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_XBOX_ONE_HOME_LED)
	HintJoystickHIDAPIGIP                  Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_GIP)
	HintJoystickHIDAPIGIPResetForMetadata  Hint = Hint(C.SDL_HINT_JOYSTICK_HIDAPI_GIP_RESET_FOR_METADATA)
	HintJoystickIokit                      Hint = Hint(C.SDL_HINT_JOYSTICK_IOKIT)
	HintJoystickLinuxClassic               Hint = Hint(C.SDL_HINT_JOYSTICK_LINUX_CLASSIC)
	HintJoystickLinuxDeadzones             Hint = Hint(C.SDL_HINT_JOYSTICK_LINUX_DEADZONES)
	HintJoystickLinuxDigitalHats           Hint = Hint(C.SDL_HINT_JOYSTICK_LINUX_DIGITAL_HATS)
	HintJoystickLinuxHatDeadzones          Hint = Hint(C.SDL_HINT_JOYSTICK_LINUX_HAT_DEADZONES)
	HintJoystickMFI                        Hint = Hint(C.SDL_HINT_JOYSTICK_MFI)
	HintJoystickRogChakram                 Hint = Hint(C.SDL_HINT_JOYSTICK_ROG_CHAKRAM)
	HintJoystickThrottleDevices            Hint = Hint(C.SDL_HINT_JOYSTICK_THROTTLE_DEVICES)
	HintJoystickThrottleDevicesExcluded    Hint = Hint(C.SDL_HINT_JOYSTICK_THROTTLE_DEVICES_EXCLUDED)
	HintJoystickWheelDevices               Hint = Hint(C.SDL_HINT_JOYSTICK_WHEEL_DEVICES)
	HintJoystickWheelDevicesExcluded       Hint = Hint(C.SDL_HINT_JOYSTICK_WHEEL_DEVICES_EXCLUDED)
	HintJoystickZeroCenteredDevices        Hint = Hint(C.SDL_HINT_JOYSTICK_ZERO_CENTERED_DEVICES)
	HintJoystickHapticAxes                 Hint = Hint(C.SDL_HINT_JOYSTICK_HAPTIC_AXES)
	HintKeycodeOptions                     Hint = Hint(C.SDL_HINT_KEYCODE_OPTIONS)
	HintKMSDRMDeviceIndex                  Hint = Hint(C.SDL_HINT_KMSDRM_DEVICE_INDEX)
	HintKMSDRMRequireDrmMaster             Hint = Hint(C.SDL_HINT_KMSDRM_REQUIRE_DRM_MASTER)
	HintKMSDRMAtomic                       Hint = Hint(C.SDL_HINT_KMSDRM_ATOMIC)
	HintLogging                            Hint = Hint(C.SDL_HINT_LOGGING)
	HintMacBackgroundApp                   Hint = Hint(C.SDL_HINT_MAC_BACKGROUND_APP)
	HintMacCtrlClickEmulateRightClick      Hint = Hint(C.SDL_HINT_MAC_CTRL_CLICK_EMULATE_RIGHT_CLICK)
	HintMacOpenGLAsyncDispatch             Hint = Hint(C.SDL_HINT_MAC_OPENGL_ASYNC_DISPATCH)
	HintMacOptionAsAlt                     Hint = Hint(C.SDL_HINT_MAC_OPTION_AS_ALT)
	HintMacScrollMomentum                  Hint = Hint(C.SDL_HINT_MAC_SCROLL_MOMENTUM)
	HintMacPressAndHold                    Hint = Hint(C.SDL_HINT_MAC_PRESS_AND_HOLD)
	HintMainCallbackRate                   Hint = Hint(C.SDL_HINT_MAIN_CALLBACK_RATE)
	HintMouseAutoCapture                   Hint = Hint(C.SDL_HINT_MOUSE_AUTO_CAPTURE)
	HintMouseDoubleClickRadius             Hint = Hint(C.SDL_HINT_MOUSE_DOUBLE_CLICK_RADIUS)
	HintMouseDoubleClickTIME               Hint = Hint(C.SDL_HINT_MOUSE_DOUBLE_CLICK_TIME)
	HintMouseDefaultSystemCursor           Hint = Hint(C.SDL_HINT_MOUSE_DEFAULT_SYSTEM_CURSOR)
	HintMouseDpiScaleCursors               Hint = Hint(C.SDL_HINT_MOUSE_DPI_SCALE_CURSORS)
	HintMouseEmulateWarpWithRelative       Hint = Hint(C.SDL_HINT_MOUSE_EMULATE_WARP_WITH_RELATIVE)
	HintMouseFocusClickThrough             Hint = Hint(C.SDL_HINT_MOUSE_FOCUS_CLICKTHROUGH)
	HintMouseNormalSpeedScale              Hint = Hint(C.SDL_HINT_MOUSE_NORMAL_SPEED_SCALE)
	HintMouseRelativeModeCenter            Hint = Hint(C.SDL_HINT_MOUSE_RELATIVE_MODE_CENTER)
	HintMouseRelativeSpeedScale            Hint = Hint(C.SDL_HINT_MOUSE_RELATIVE_SPEED_SCALE)
	HintMouseRelativeSystemScale           Hint = Hint(C.SDL_HINT_MOUSE_RELATIVE_SYSTEM_SCALE)
	HintMouseRelativeWarpMotion            Hint = Hint(C.SDL_HINT_MOUSE_RELATIVE_WARP_MOTION)
	HintMouseRelativeCursorVisible         Hint = Hint(C.SDL_HINT_MOUSE_RELATIVE_CURSOR_VISIBLE)
	HintMouseTouchEvents                   Hint = Hint(C.SDL_HINT_MOUSE_TOUCH_EVENTS)
	HintMuteConsoleKeyboard                Hint = Hint(C.SDL_HINT_MUTE_CONSOLE_KEYBOARD)
	HintNoSignalHandlers                   Hint = Hint(C.SDL_HINT_NO_SIGNAL_HANDLERS)
	HintOpenGLLibrary                      Hint = Hint(C.SDL_HINT_OPENGL_LIBRARY)
	HintEglLibrary                         Hint = Hint(C.SDL_HINT_EGL_LIBRARY)
	HintOpenGLEsDriver                     Hint = Hint(C.SDL_HINT_OPENGL_ES_DRIVER)
	HintOpenGLForceSRGBFramebuffer         Hint = Hint(C.SDL_HINT_OPENGL_FORCE_SRGB_FRAMEBUFFER)
	HintOpenvrLibrary                      Hint = Hint(C.SDL_HINT_OPENVR_LIBRARY)
	HintOrientations                       Hint = Hint(C.SDL_HINT_ORIENTATIONS)
	HintPollSentinel                       Hint = Hint(C.SDL_HINT_POLL_SENTINEL)
	HintPreferredLocales                   Hint = Hint(C.SDL_HINT_PREFERRED_LOCALES)
	HintQuitOnLastWindowClose              Hint = Hint(C.SDL_HINT_QUIT_ON_LAST_WINDOW_CLOSE)
	HintRenderDirect3dThreadsafe           Hint = Hint(C.SDL_HINT_RENDER_DIRECT3D_THREADSAFE)
	HintRenderDirect3d11Debug              Hint = Hint(C.SDL_HINT_RENDER_DIRECT3D11_DEBUG)
	HintRenderDirect3d11Warp               Hint = Hint(C.SDL_HINT_RENDER_DIRECT3D11_WARP)
	HintRenderVulkanDebug                  Hint = Hint(C.SDL_HINT_RENDER_VULKAN_DEBUG)
	HintRenderGpuDebug                     Hint = Hint(C.SDL_HINT_RENDER_GPU_DEBUG)
	HintRenderGpuLowPower                  Hint = Hint(C.SDL_HINT_RENDER_GPU_LOW_POWER)
	HintRenderDriver                       Hint = Hint(C.SDL_HINT_RENDER_DRIVER)
	HintRenderLineMethod                   Hint = Hint(C.SDL_HINT_RENDER_LINE_METHOD)
	HintRenderMetalPreferLowPowerDevice    Hint = Hint(C.SDL_HINT_RENDER_METAL_PREFER_LOW_POWER_DEVICE)
	HintRenderVsync                        Hint = Hint(C.SDL_HINT_RENDER_VSYNC)
	HintReturnKeyHidesIME                  Hint = Hint(C.SDL_HINT_RETURN_KEY_HIDES_IME)
	HintRogGamepadMice                     Hint = Hint(C.SDL_HINT_ROG_GAMEPAD_MICE)
	HintRogGamepadMiceExcluded             Hint = Hint(C.SDL_HINT_ROG_GAMEPAD_MICE_EXCLUDED)
	HintPS2GSWidth                         Hint = Hint(C.SDL_HINT_PS2_GS_WIDTH)
	HintPS2GSHeight                        Hint = Hint(C.SDL_HINT_PS2_GS_HEIGHT)
	HintPS2GSProgressive                   Hint = Hint(C.SDL_HINT_PS2_GS_PROGRESSIVE)
	HintPS2GSMode                          Hint = Hint(C.SDL_HINT_PS2_GS_MODE)
	HintRpiVideoLayer                      Hint = Hint(C.SDL_HINT_RPI_VIDEO_LAYER)
	HintScreensaverInhibitActivityName     Hint = Hint(C.SDL_HINT_SCREENSAVER_INHIBIT_ACTIVITY_NAME)
	HintShutdownDBusOnQuit                 Hint = Hint(C.SDL_HINT_SHUTDOWN_DBUS_ON_QUIT)
	HintStorageTitleDriver                 Hint = Hint(C.SDL_HINT_STORAGE_TITLE_DRIVER)
	HintStorageUserDriver                  Hint = Hint(C.SDL_HINT_STORAGE_USER_DRIVER)
	HintThreadForceRealtimeTimeCritical    Hint = Hint(C.SDL_HINT_THREAD_FORCE_REALTIME_TIME_CRITICAL)
	HintThreadPriorityPolicy               Hint = Hint(C.SDL_HINT_THREAD_PRIORITY_POLICY)
	HintTimerResolution                    Hint = Hint(C.SDL_HINT_TIMER_RESOLUTION)
	HintTouchMouseEvents                   Hint = Hint(C.SDL_HINT_TOUCH_MOUSE_EVENTS)
	HintTrackpadIsTouchOnly                Hint = Hint(C.SDL_HINT_TRACKPAD_IS_TOUCH_ONLY)
	HintTvRemoteAsJoystick                 Hint = Hint(C.SDL_HINT_TV_REMOTE_AS_JOYSTICK)
	HintVideoAllowScreensaver              Hint = Hint(C.SDL_HINT_VIDEO_ALLOW_SCREENSAVER)
	HintVideoDisplayPriority               Hint = Hint(C.SDL_HINT_VIDEO_DISPLAY_PRIORITY)
	HintVideoDoubleBuffer                  Hint = Hint(C.SDL_HINT_VIDEO_DOUBLE_BUFFER)
	HintVideoDriver                        Hint = Hint(C.SDL_HINT_VIDEO_DRIVER)
	HintVideoDummySaveFrames               Hint = Hint(C.SDL_HINT_VIDEO_DUMMY_SAVE_FRAMES)
	HintVideoEglAllowGetdisplayFallback    Hint = Hint(C.SDL_HINT_VIDEO_EGL_ALLOW_GETDISPLAY_FALLBACK)
	HintVideoForceEGL                      Hint = Hint(C.SDL_HINT_VIDEO_FORCE_EGL)
	HintVideoMacFullscreenSpaces           Hint = Hint(C.SDL_HINT_VIDEO_MAC_FULLSCREEN_SPACES)
	HintVideoMacFullscreenMenuVisibility   Hint = Hint(C.SDL_HINT_VIDEO_MAC_FULLSCREEN_MENU_VISIBILITY)
	HintVideoMetalAutoResizeDrawable       Hint = Hint(C.SDL_HINT_VIDEO_METAL_AUTO_RESIZE_DRAWABLE)
	HintVideoMatchExclusiveModeOnMove      Hint = Hint(C.SDL_HINT_VIDEO_MATCH_EXCLUSIVE_MODE_ON_MOVE)
	HintVideoMinimizeOnFocusLoss           Hint = Hint(C.SDL_HINT_VIDEO_MINIMIZE_ON_FOCUS_LOSS)
	HintVideoOffscreenSaveFrames           Hint = Hint(C.SDL_HINT_VIDEO_OFFSCREEN_SAVE_FRAMES)
	HintVideoSyncWindowOperations          Hint = Hint(C.SDL_HINT_VIDEO_SYNC_WINDOW_OPERATIONS)
	HintVideoWaylandAllowLibdecor          Hint = Hint(C.SDL_HINT_VIDEO_WAYLAND_ALLOW_LIBDECOR)
	HintVideoWaylandModeEmulation          Hint = Hint(C.SDL_HINT_VIDEO_WAYLAND_MODE_EMULATION)
	HintVideoWaylandModeScaling            Hint = Hint(C.SDL_HINT_VIDEO_WAYLAND_MODE_SCALING)
	HintVideoWaylandPreferLibdecor         Hint = Hint(C.SDL_HINT_VIDEO_WAYLAND_PREFER_LIBDECOR)
	HintVideoWaylandScaleToDisplay         Hint = Hint(C.SDL_HINT_VIDEO_WAYLAND_SCALE_TO_DISPLAY)
	HintVideoWinD3dcompiler                Hint = Hint(C.SDL_HINT_VIDEO_WIN_D3DCOMPILER)
	HintVideoX11EnableXSyncExt             Hint = Hint(C.SDL_HINT_VIDEO_X11_ENABLE_XSYNC_EXT)
	HintVideoX11ExternalWindowInput        Hint = Hint(C.SDL_HINT_VIDEO_X11_EXTERNAL_WINDOW_INPUT)
	HintVideoX11NetWmBypassCompositor      Hint = Hint(C.SDL_HINT_VIDEO_X11_NET_WM_BYPASS_COMPOSITOR)
	HintVideoX11NetWmPing                  Hint = Hint(C.SDL_HINT_VIDEO_X11_NET_WM_PING)
	HintVideoX11Nodirectcolor              Hint = Hint(C.SDL_HINT_VIDEO_X11_NODIRECTCOLOR)
	HintVideoX11ScalingFactor              Hint = Hint(C.SDL_HINT_VIDEO_X11_SCALING_FACTOR)
	HintVideoX11VisualID                   Hint = Hint(C.SDL_HINT_VIDEO_X11_VISUALID)
	HintVideoX11WindowVisualID             Hint = Hint(C.SDL_HINT_VIDEO_X11_WINDOW_VISUALID)
	HintVideoX11XRandR                     Hint = Hint(C.SDL_HINT_VIDEO_X11_XRANDR)
	HintVitaEnableBackTouch                Hint = Hint(C.SDL_HINT_VITA_ENABLE_BACK_TOUCH)
	HintVitaEnableFrontTouch               Hint = Hint(C.SDL_HINT_VITA_ENABLE_FRONT_TOUCH)
	HintVitaModulePath                     Hint = Hint(C.SDL_HINT_VITA_MODULE_PATH)
	HintVitaPVRInit                        Hint = Hint(C.SDL_HINT_VITA_PVR_INIT)
	HintVitaResolution                     Hint = Hint(C.SDL_HINT_VITA_RESOLUTION)
	HintVitaPVROpenGL                      Hint = Hint(C.SDL_HINT_VITA_PVR_OPENGL)
	HintVitaTouchMouseDevice               Hint = Hint(C.SDL_HINT_VITA_TOUCH_MOUSE_DEVICE)
	HintVulkanDisplay                      Hint = Hint(C.SDL_HINT_VULKAN_DISPLAY)
	HintVulkanLibrary                      Hint = Hint(C.SDL_HINT_VULKAN_LIBRARY)
	HintWaveFactChunk                      Hint = Hint(C.SDL_HINT_WAVE_FACT_CHUNK)
	HintWaveChunkLimit                     Hint = Hint(C.SDL_HINT_WAVE_CHUNK_LIMIT)
	HintWaveRIFFChunkSize                  Hint = Hint(C.SDL_HINT_WAVE_RIFF_CHUNK_SIZE)
	HintWaveTruncation                     Hint = Hint(C.SDL_HINT_WAVE_TRUNCATION)
	HintWindowActivateWhenRaised           Hint = Hint(C.SDL_HINT_WINDOW_ACTIVATE_WHEN_RAISED)
	HintWindowActivateWhenShown            Hint = Hint(C.SDL_HINT_WINDOW_ACTIVATE_WHEN_SHOWN)
	HintWindowAllowTopmost                 Hint = Hint(C.SDL_HINT_WINDOW_ALLOW_TOPMOST)
	HintWindowFrameUsableWhileCursorHidden Hint = Hint(C.SDL_HINT_WINDOW_FRAME_USABLE_WHILE_CURSOR_HIDDEN)
	HintWindowsCloseOnAltF4                Hint = Hint(C.SDL_HINT_WINDOWS_CLOSE_ON_ALT_F4)
	HintWindowsEnableMenuMnemonics         Hint = Hint(C.SDL_HINT_WINDOWS_ENABLE_MENU_MNEMONICS)
	HintWindowsEnableMessageloop           Hint = Hint(C.SDL_HINT_WINDOWS_ENABLE_MESSAGELOOP)
	HintWindowsGameinput                   Hint = Hint(C.SDL_HINT_WINDOWS_GAMEINPUT)
	HintWindowsRawKeyboard                 Hint = Hint(C.SDL_HINT_WINDOWS_RAW_KEYBOARD)
	HintWindowsRawKeyboardExcludeHotkeys   Hint = Hint(C.SDL_HINT_WINDOWS_RAW_KEYBOARD_EXCLUDE_HOTKEYS)
	HintWindowsRawKeyboardInputsink        Hint = Hint(C.SDL_HINT_WINDOWS_RAW_KEYBOARD_INPUTSINK)
	HintWindowsForceSemaphoreKernel        Hint = Hint(C.SDL_HINT_WINDOWS_FORCE_SEMAPHORE_KERNEL)
	HintWindowsIntResourceIcon             Hint = Hint(C.SDL_HINT_WINDOWS_INTRESOURCE_ICON)
	HintWindowsIntResourceIconSmall        Hint = Hint(C.SDL_HINT_WINDOWS_INTRESOURCE_ICON_SMALL)
	HintWindowsUseD3d9ex                   Hint = Hint(C.SDL_HINT_WINDOWS_USE_D3D9EX)
	HintWindowsEraseBackgroundMode         Hint = Hint(C.SDL_HINT_WINDOWS_ERASE_BACKGROUND_MODE)
	HintX11ForceOverrideRedirect           Hint = Hint(C.SDL_HINT_X11_FORCE_OVERRIDE_REDIRECT)
	HintX11WindowType                      Hint = Hint(C.SDL_HINT_X11_WINDOW_TYPE)
	HintX11XcbLibrary                      Hint = Hint(C.SDL_HINT_X11_XCB_LIBRARY)
	HintXInputEnabled                      Hint = Hint(C.SDL_HINT_XINPUT_ENABLED)
	HintAssert                             Hint = Hint(C.SDL_HINT_ASSERT)
	HintPenMouseEvents                     Hint = Hint(C.SDL_HINT_PEN_MOUSE_EVENTS)
	HintPenTouchEvents                     Hint = Hint(C.SDL_HINT_PEN_TOUCH_EVENTS)
)

// SetHint sets a hint with normal priority.
//
// Hints will not be set if there is an existing override hint or environment
// variable that takes precedence. You can use SDL_SetHintWithPriority() to
// set the hint with override priority instead.
//
// \param name the hint to set.
// \param value the value of the hint variable.
// \returns true on success or false on failure; call SDL_GetError() for more
//
//	         information.
//
//	hreadsafety It is safe to call this function from any thread.
//
// \since This function is available since SDL 3.2.0.
//
// \sa SDL_GetHint
// \sa SDL_ResetHint
// \sa SDL_SetHintWithPriority
func SetHint(name Hint, value string) error {
	cName := C.CString(string(name))
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	if !C.SDL_SetHint(cName, cValue) {
		return GetError()
	}
	return nil
}

// SetHintWithPriority sets a hint with a specific priority.
//
// The priority controls the behavior when setting a hint that already has a
// value. Hints will replace existing hints of their priority and lower.
// Environment variables are considered to have override priority.
//
// \param name the hint to set.
// \param value the value of the hint variable.
// \param priority the SDL_HintPriority level for the hint.
// \returns true on success or false on failure; call SDL_GetError() for more
//
//	         information.
//
//	hreadsafety It is safe to call this function from any thread.
//
// \since This function is available since SDL 3.2.0.
//
// \sa SDL_GetHint
// \sa SDL_ResetHint
// \sa SDL_SetHint
func SetHintWithPriority(name Hint, value string, priority HintPriority) error {
	cName := C.CString(string(name))
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	if !C.SDL_SetHintWithPriority(cName, cValue, C.SDL_HintPriority(priority)) {
		return GetError()
	}
	return nil
}

// GetHint gets the value of a hint.
//
// \param name the hint to query.
// \returns the string value of a hint or NULL if the hint isn't set.
//
//	hreadsafety It is safe to call this function from any thread.
//
// \since This function is available since SDL 3.2.0.
//
// \sa SDL_SetHint
// \sa SDL_SetHintWithPriority
func GetHint(name Hint) string {
	cName := C.CString(string(name))
	defer C.free(unsafe.Pointer(cName))

	value := C.SDL_GetHint(cName)
	if value == nil {
		return ""
	}
	return C.GoString(value)
}

// GetHintBoolean gets the boolean value of a hint variable.
//
// \param name the name of the hint to get the boolean value from.
// \param default_value the value to return if the hint does not exist.
// \returns the boolean value of a hint or the provided default value if the
//
//	         hint does not exist.
//
//	hreadsafety It is safe to call this function from any thread.
//
// \since This function is available since SDL 3.2.0.
//
// \sa SDL_GetHint
// \sa SDL_SetHint
func GetHintBoolean(name Hint, defaultValue bool) bool {
	cName := C.CString(string(name))
	defer C.free(unsafe.Pointer(cName))

	return bool(C.SDL_GetHintBoolean(cName, C.bool(defaultValue)))
}

// ResetHint resets a hint to the default value.
//
// This will reset a hint to the value of the environment variable, or NULL if
// the environment isn't set. Callbacks will be called normally with this
// change.
//
// \param name the hint to set.
// \returns true on success or false on failure; call SDL_GetError() for more
//
//	         information.
//
//	hreadsafety It is safe to call this function from any thread.
//
// \since This function is available since SDL 3.2.0.
//
// \sa SDL_SetHint
// \sa SDL_ResetHints
func ResetHint(name Hint) {
	cName := C.CString(string(name))
	defer C.free(unsafe.Pointer(cName))

	C.SDL_ResetHint(cName)
}

// ResetHints resets all hints to the default values.
//
// This will reset all hints to the value of the associated environment
// variable, or NULL if the environment isn't set. Callbacks will be called
// normally with this change.
//
//	hreadsafety It is safe to call this function from any thread.
//
// \since This function is available since SDL 3.2.0.
//
// \sa SDL_ResetHint
func ResetHints() {
	C.SDL_ResetHints()
}
