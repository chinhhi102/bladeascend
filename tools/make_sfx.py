#!/usr/bin/env python3
"""Synthesise placeholder SFX with ffmpeg (self-made, CC0) into assets/src/audio/sfx/.

Run from the repo root:
    python3 tools/make_sfx.py              # only sounds that are missing
    python3 tools/make_sfx.py --all        # rebuild every sound
    python3 tools/make_sfx.py skill_loa    # just these names
Replace any file with a real sound later (keep the same name); see assets/src/CREDITS.md.
"""
import pathlib
import subprocess
import sys

OUT = pathlib.Path(__file__).resolve().parent.parent / "assets" / "src" / "audio" / "sfx"
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
    # Kỹ năng (spec v6.36 §6.5 lớp ① và ⑤) — tiếng tạm, thay bằng tiếng thật cùng tên sau.
    "skill_banh_bao": ("sin(2*PI*(220+80*sin(2*PI*7*t))*t)*exp(-t*5)*0.8+0.3*sin(2*PI*90*t)*between(t,0.35,0.6)*exp(-(t-0.35)*8)", 0.7, 0.7),
    "skill_loa": (f"((2*gte(sin(2*PI*440*t),0)-1)*0.35+0.3*sin(2*PI*880*t))*(1+0.5*sin(2*PI*12*t))*exp(-t*2.5)+0.15*{N}*exp(-t*6)", 0.6, 0.6),
    "skill_hat_ru": ("sin(2*PI*(392*lt(t,0.25)+330*between(t,0.25,0.5)+294*gte(t,0.5))*t)*(1+0.2*sin(2*PI*5*t))*exp(-mod(t,0.25)*3)*0.7", 0.75, 0.5),
    "skill_sleep": ("sin(2*PI*(300-150*t)*t)*exp(-t*3)*0.6+0.3*sin(2*PI*600*t)*exp(-t*9)", 0.6, 0.5),
    "skill_dan_em": ("sin(2*PI*(1800+600*sin(2*PI*3*t))*t)*between(t,0,0.35)*exp(-t*3)+sin(2*PI*(2200-400*t)*t)*between(t,0.4,0.7)*0.8", 0.75, 0.5),
    "skill_tran_chau": (f"0.5*{N}*exp(-t*8)*(0.6+0.4*sin(2*PI*40*t))+0.5*sin(2*PI*(500-300*t)*t)*exp(-t*10)", 0.45, 0.6),
    "skill_bay_hang": (f"0.6*sin(2*PI*140*t)*exp(-mod(t,0.1)*30)*lt(t,0.5)+0.3*{N}*exp(-mod(t,0.1)*40)*lt(t,0.5)", 0.55, 0.7),
    "skill_hut_mi": (f"0.4*{N}*(0.5+0.5*sin(2*PI*30*t))*exp(-t*3)+0.5*sin(2*PI*(900-700*t)*t)*exp(-t*4)", 0.6, 0.6),
    "skill_gong_co": ("sin(2*PI*(90+60*t)*t)*(t/0.3)*lt(t,0.3)+sin(2*PI*150*t)*gte(t,0.3)*exp(-(t-0.3)*6)", 0.6, 0.9),
    "skill_gom_don": (f"0.6*sin(2*PI*(300+2400*t)*t)*exp(-t*3)+0.3*{N}*exp(-t*5)", 0.5, 0.6),
    "skill_rut_co": ("sin(2*PI*(260-180*t)*t)*exp(-t*6)*0.8+0.3*sin(2*PI*520*t)*exp(-t*14)", 0.4, 0.7),
    "skill_end": (f"0.5*sin(2*PI*(700-500*t)*t)*exp(-t*9)+0.2*{N}*exp(-t*14)", 0.3, 0.4),
}


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    args = sys.argv[1:]
    rebuild = "--all" in args
    only = {a for a in args if not a.startswith("--")}
    for name, (expr, dur, vol) in SOUNDS.items():
        dst = OUT / f"{name}.ogg"
        if only and name not in only:
            continue
        if not only and not rebuild and dst.exists():
            continue
        subprocess.run([
            "ffmpeg", "-y", "-loglevel", "error", "-f", "lavfi",
            "-i", f"aevalsrc=exprs='{expr}':s=44100:d={dur}",
            "-af", f"volume={vol},afade=t=out:st={max(dur - 0.05, 0)}:d=0.05,alimiter=limit=0.9",
            "-ac", "1", "-c:a", "libvorbis", "-q:a", "3", str(dst),
        ], check=True)
        print("wrote", dst.name)


if __name__ == "__main__":
    main()
