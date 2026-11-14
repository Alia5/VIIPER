namespace Viiper.Client.Devices.Dualsense;

public class TriggerEffect
{
    private const int BitsPerStrength = 3;
    private const int StrengthMask = 0x07;
    private const int FrequencyParam = 8;

    public byte Mode { get; init; }
    public byte[] Params { get; init; } = new byte[Dualsense.TriggerParamsSize];
    public byte[] ZoneStrengths { get; init; } = new byte[Dualsense.TriggerZoneCount];
    public byte StartZone { get; private set; }
    public byte EndZone { get; private set; }
    public byte StartPosition { get; private set; }
    public byte EndPosition { get; private set; }
    public byte Strength { get; private set; }
    public byte SnapForce { get; private set; }
    public byte AmplitudeA { get; private set; }
    public byte AmplitudeB { get; private set; }
    public byte FirstFoot { get; private set; }
    public byte SecondFoot { get; private set; }
    public byte Frequency { get; private set; }
    public byte Period { get; private set; }

    public static TriggerEffect Decode(byte mode, byte[] parameters)
    {
        var effect = new TriggerEffect
        {
            Mode = mode,
            Params = parameters,
        };
        var zones = parameters[0] | parameters[1] << 8;

        switch ((TriggerMode)mode)
        {
            case TriggerMode.Feedback:
            case TriggerMode.Vibration:
                var packedStrengths = (uint)(parameters[2] | parameters[3] << 8 | parameters[4] << 16 | parameters[5] << 24);
                for (var zone = 0; zone < Dualsense.TriggerZoneCount; zone++)
                {
                    if ((zones & 1 << zone) != 0)
                    {
                        effect.ZoneStrengths[zone] = (byte)((packedStrengths >> BitsPerStrength * zone & StrengthMask) + 1);
                    }
                }
                if ((TriggerMode)mode == TriggerMode.Vibration)
                {
                    effect.Frequency = parameters[FrequencyParam];
                }
                break;
            case TriggerMode.Weapon:
            case TriggerMode.Bow:
            case TriggerMode.Galloping:
            case TriggerMode.Machine:
                for (var zone = Dualsense.TriggerZoneCount - 1; zone >= 0; zone--)
                {
                    if ((zones & 1 << zone) != 0)
                    {
                        effect.StartZone = (byte)zone;
                    }
                }
                for (var zone = 0; zone < Dualsense.TriggerZoneCount; zone++)
                {
                    if ((zones & 1 << zone) != 0)
                    {
                        effect.EndZone = (byte)zone;
                    }
                }
                var first = (byte)(parameters[2] & StrengthMask);
                var second = (byte)(parameters[2] >> BitsPerStrength & StrengthMask);
                switch ((TriggerMode)mode)
                {
                    case TriggerMode.Weapon:
                        effect.Strength = (byte)(first + 1);
                        break;
                    case TriggerMode.Bow:
                        effect.Strength = (byte)(first + 1);
                        effect.SnapForce = (byte)(second + 1);
                        break;
                    case TriggerMode.Galloping:
                        effect.SecondFoot = first;
                        effect.FirstFoot = second;
                        effect.Frequency = parameters[3];
                        break;
                    case TriggerMode.Machine:
                        effect.AmplitudeA = first;
                        effect.AmplitudeB = second;
                        effect.Frequency = parameters[3];
                        effect.Period = parameters[4];
                        break;
                }
                break;
            case TriggerMode.SimpleFeedback:
            case TriggerMode.LimitedFeedback:
                effect.StartPosition = parameters[0];
                effect.Strength = parameters[1];
                break;
            case TriggerMode.SimpleWeapon:
            case TriggerMode.LimitedWeapon:
                effect.StartPosition = parameters[0];
                effect.EndPosition = parameters[1];
                effect.Strength = parameters[2];
                break;
            case TriggerMode.SimpleVibration:
                effect.Frequency = parameters[0];
                effect.Strength = parameters[1];
                effect.StartPosition = parameters[2];
                break;
        }
        return effect;
    }
}
