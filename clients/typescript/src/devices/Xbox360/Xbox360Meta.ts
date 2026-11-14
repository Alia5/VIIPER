// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


// Typed deviceSpecific helpers for Xbox360


export interface Xbox360CreateOptions {
  subType?: number;
}

export const xbox360CreateOptionsToMap = (value: Xbox360CreateOptions): Record<string, unknown> => ({
  subType: value.subType,
});

export const xbox360CreateOptionsFromMap = (data: Record<string, unknown> | undefined | null): Xbox360CreateOptions => {
  const map = data ?? {};
  return {
    subType: map['subType'] as number,
  };
};


