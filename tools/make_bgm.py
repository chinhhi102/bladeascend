#!/usr/bin/env python3
"""Synthesise a placeholder looping BGM (fast pentatonic, plucked-string feel) → assets/audio/bgm/main.ogg.

Run from the repo root: python3 tools/make_bgm.py   (needs ffmpeg for the .ogg encode)
"""
import math
import pathlib
import random
import struct
import subprocess
import wave

RATE = 22050
BPM = 150
BEAT = 60 / BPM
OUT = pathlib.Path(__file__).resolve().parent.parent / "assets" / "audio" / "bgm"

# D major pentatonic (D E F# A B) across two octaves, as MIDI notes.
SCALE = [62, 64, 66, 69, 71, 74, 76, 78, 81, 83]
BASS = [38, 45, 43, 45]  # D A G A, one per bar


def freq(midi: int) -> float:
    return 440 * 2 ** ((midi - 69) / 12)


def pluck(buf: list, start: float, midi: int, dur: float, vol: float) -> None:
    f = freq(midi)
    i0 = int(start * RATE)
    for i in range(int(dur * RATE)):
        t = i / RATE
        env = math.exp(-t * 7)
        s = math.sin(2 * math.pi * f * t) + 0.35 * math.sin(4 * math.pi * f * t) * math.exp(-t * 12)
        if i0 + i < len(buf):
            buf[i0 + i] += vol * env * s


def drum(buf: list, start: float, low: bool) -> None:
    i0 = int(start * RATE)
    for i in range(int(0.12 * RATE)):
        t = i / RATE
        s = math.sin(2 * math.pi * (70 + 120 * math.exp(-t * 40)) * t) if low else random.uniform(-1, 1)
        if i0 + i < len(buf):
            buf[i0 + i] += (0.5 if low else 0.12) * math.exp(-t * (25 if low else 60)) * s


def main() -> None:
    random.seed(7)
    bars = 8
    total = bars * 4 * BEAT
    buf = [0.0] * int(total * RATE + RATE)
    idx = 4
    for bar in range(bars):
        pluck(buf, bar * 4 * BEAT, BASS[bar % 4], 4 * BEAT, 0.35)
        for step in range(8):  # eighth notes
            t = (bar * 8 + step) * BEAT / 2
            idx = max(0, min(len(SCALE) - 1, idx + random.choice([-2, -1, -1, 1, 1, 2, 0])))
            if random.random() < 0.85:
                pluck(buf, t, SCALE[idx], BEAT, 0.22)
            drum(buf, t, step % 4 == 0)
    # Wrap the tail into the start so the loop is seamless.
    n = int(total * RATE)
    for i in range(len(buf) - n):
        buf[i] += buf[n + i]
    buf = buf[:n]
    peak = max(abs(x) for x in buf) or 1
    OUT.mkdir(parents=True, exist_ok=True)
    wav = OUT / "main.wav"
    with wave.open(str(wav), "wb") as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(RATE)
        w.writeframes(b"".join(struct.pack("<h", int(x / peak * 0.8 * 32767)) for x in buf))
    subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-i", str(wav), "-c:a", "libvorbis", "-q:a", "3", str(OUT / "main.ogg")], check=True)
    wav.unlink()
    print("wrote", OUT / "main.ogg")


if __name__ == "__main__":
    main()
