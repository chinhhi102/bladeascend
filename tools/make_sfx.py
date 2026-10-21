#!/usr/bin/env python3
"""Synthesise placeholder SFX with ffmpeg (self-made, CC0) into assets/audio/sfx/.

Run from the repo root: python3 tools/make_sfx.py
Replace any file with a real sound later (keep the same name); see assets/CREDITS.md.
"""
import pathlib
import subprocess

OUT = pathlib.Path(__file__).resolve().parent.parent / "assets" / "audio" / "sfx"
N = "(random(0)*2-1)"  # white noise

# name: (expression of t, duration seconds, volume)
SOUNDS = {
    "sword_clash": (f"(sin(2*PI*2300*t)+0.7*sin(2*PI*3470*t)+0.5*sin(2*PI*5110*t))*exp(-t*14)/2.2+0.3*{N}*exp(-t*60)", 0.45, 0.8),
    "sword_hit_body": (f"sin(2*PI*(110+260*exp(-t*30))*t)*exp(-t*12)+0.45*{N}*exp(-t*45)", 0.3, 0.9),
    "pickup_sword": ("sin(2*PI*(700+1800*t)*t)*exp(-t*7)", 0.3, 0.6),
    "pickup_buff": ("(sin(2*PI*880*t)*lt(t,0.09)+sin(2*PI*1320*t)*gte(t,0.09))*exp(-t*6)", 0.3, 0.6),
    "pickup_exp": ("sin(2*PI*1760*t)*exp(-t*16)+0.3*sin(2*PI*3520*t)*exp(-t*22)", 0.2, 0.5),
    "buff_end": (f"0.7*sin(2*PI*(520-500*t)*t)*exp(-t*3)+0.25*{N}*exp(-t*6)", 0.45, 0.5),
    "sword_break": (f"0.6*{N}*exp(-t*9)+0.5*(sin(2*PI*3900*t)+sin(2*PI*5300*t))*exp(-t*11)/2", 0.5, 0.8),
    "shield_break": (f"0.5*{N}*exp(-t*7)+0.6*(sin(2*PI*1500*t)+sin(2*PI*2250*t))*exp(-t*8)/2", 0.6, 0.8),
    "chest_incoming": ("sin(2*PI*70*t)*exp(-mod(t,0.25)*16)+0.4*sin(2*PI*140*t)*exp(-mod(t,0.25)*25)", 1.0, 0.9),
    "chest_open": ("sin(2*PI*(523*lt(t,0.1)+659*between(t,0.1,0.2)+784*between(t,0.2,0.3)+1047*gte(t,0.3))*t)*exp(-mod(t,0.1)*6)*exp(-t*2)", 0.7, 0.6),
    "obstacle_hit": (f"sin(2*PI*320*t)*exp(-t*35)+0.5*{N}*exp(-t*70)", 0.15, 0.6),
    "obstacle_break": (f"{N}*exp(-t*5)*(0.6+0.4*sin(2*PI*9*t))+0.6*sin(2*PI*60*t)*exp(-t*6)", 0.8, 0.8),
    "obstacle_regrow": ("sin(2*PI*(180+900*t)*t)*exp(-t*9)", 0.25, 0.6),
    "gate_in": (f"{N}*(t/0.4)*exp(-t*2)*0.7+0.5*sin(2*PI*(300+1500*t)*t)*exp(-t*4)", 0.45, 0.6),
    "gate_out": (f"sin(2*PI*(90-50*t)*t)*exp(-t*7)+0.4*{N}*exp(-t*20)", 0.45, 0.8),
    "gate_warn": ("sin(2*PI*1050*t)*lt(mod(t,0.12),0.06)*exp(-t*2)", 0.5, 0.4),
    "monster_hurt": ("sin(2*PI*(950-600*t)*t)*exp(-t*12)", 0.2, 0.5),
    "monster_pop": (f"sin(2*PI*(520+1400*exp(-t*40))*t)*exp(-t*18)+0.3*{N}*exp(-t*50)", 0.2, 0.6),
    "levelup": ("(sin(2*PI*1319*t)+0.8*sin(2*PI*1568*t)*gte(t,0.08))*exp(-t*4)/1.8", 0.55, 0.6),
    "breakthrough": (f"0.7*{N}*exp(-t*2.5)*(0.5+0.5*sin(2*PI*5*t))+0.5*(sin(2*PI*262*t)+sin(2*PI*330*t)+sin(2*PI*392*t))*gte(t,0.3)*exp(-(t-0.3)*1.5)/3", 1.6, 0.9),
    "realm_drop": ("sin(2*PI*(392*lt(t,0.35)+370*between(t,0.35,0.7)+349*between(t,0.7,1.05)+311*gte(t,1.05))*(1+0.01*sin(2*PI*6*t))*t)*exp(-mod(t,0.35)*1.5)*0.8", 1.6, 0.6),
    "death": (f"0.6*sin(2*PI*(160+300*exp(-t*25))*t)*exp(-t*10)+0.5*sin(2*PI*(1400-1100*t)*t)*between(t,0.15,0.9)", 0.9, 0.7),
    "respawn": (f"0.5*{N}*between(t,0,0.4)*(t/0.4)+0.8*sin(2*PI*(80+120*exp(-(t-0.4)*20))*t)*gte(t,0.4)*exp(-(t-0.4)*10)", 0.7, 0.8),
}


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    for name, (expr, dur, vol) in SOUNDS.items():
        dst = OUT / f"{name}.ogg"
        subprocess.run([
            "ffmpeg", "-y", "-loglevel", "error", "-f", "lavfi",
            "-i", f"aevalsrc=exprs='{expr}':s=44100:d={dur}",
            "-af", f"volume={vol},afade=t=out:st={max(dur - 0.05, 0)}:d=0.05,alimiter=limit=0.9",
            "-ac", "1", "-c:a", "libvorbis", "-q:a", "3", str(dst),
        ], check=True)
        print("wrote", dst.name)


if __name__ == "__main__":
    main()
