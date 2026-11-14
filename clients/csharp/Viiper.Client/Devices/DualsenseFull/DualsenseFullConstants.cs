using System.Collections.Generic;
using Viiper.Client.Devices.Dualsense;

namespace Viiper.Client.Devices.DualsenseFull;

public static class DualsenseFull
{
    public const string DeviceType = "dualsense_full";
    public const string DeviceTypeEdge = "dualsenseedge_full";

    public const int OutputSize = Dualsense.Dualsense.OutputSize + 6;
    public const byte PowerSaveControlMicrophoneMute = 0x10;
    public const byte AudioControlOutputPath = 0x30;

    public const int OutputReportOffsetSpeakerVolume = 6;
    public const int OutputReportOffsetMicrophoneVolume = 7;
    public const int OutputReportOffsetAudioControl = 8;
    public const int OutputReportOffsetMicMuteLed = 9;
    public const int OutputReportOffsetPowerSaveControl = 10;
    public const int OutputReportOffsetAudioControl2 = 38;

    public const int AudioSampleRate = 48000;
    public const int AudioBytesPerSample = 2;
    public const int AudioOutputChannels = 4;
    public const int AudioMicrophoneChannels = 2;
    public const int AudioOutputPacketSize = AudioSampleRate / 1000 * AudioOutputChannels * AudioBytesPerSample;
    public const int AudioMicrophonePacketSize = AudioSampleRate / 1000 * AudioMicrophoneChannels * AudioBytesPerSample;
}

public class CreateOptions
{
    public MetaState? Meta { get; set; }
    public ushort? AudioPort { get; set; }

    public Dictionary<string, object?> ToMap()
    {
        var map = new Dictionary<string, object?>();
        if (Meta != null)
        {
            map["Meta"] = Meta.ToMap();
        }
        if (AudioPort != null)
        {
            map["audio_port"] = AudioPort;
        }
        return map;
    }
}
