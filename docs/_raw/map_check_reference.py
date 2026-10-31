import json,glob,math
import numpy as np
from collections import deque
G=50; N=12000//G
def load(f):
    d=json.load(open(f))
    L={l['name']:l['objects'] for l in d['layers'] if l['type']=='objectgroup'}
    return d,L
def raster(obs, r):
    occ=np.zeros((N,N),bool)
    ys,xs=np.mgrid[0:N,0:N]; cx=(xs+0.5)*G; cy=(ys+0.5)*G
    for o in obs:
        x,y,w,h=o['x'],o['y'],o['width'],o['height']
        if o.get('ellipse'):
            ox,oy,rr=x+w/2,y+h/2,max(w,h)/2
            m=(cx-ox)**2+(cy-oy)**2 <= (rr+r)**2
        else:
            dx=np.maximum(np.maximum(x-cx,0),cx-(x+w)); dy=np.maximum(np.maximum(y-cy,0),cy-(y+h))
            m=dx*dx+dy*dy <= r*r
        occ|=m
    # play margin 420
    mg=420+r
    occ|=(cx<mg)|(cx>12000-mg)|(cy<mg)|(cy>12000-mg)
    return occ
def comps(occ):
    lab=-np.ones(occ.shape,int); sizes=[]
    for sy in range(N):
        for sx in range(N):
            if occ[sy,sx] or lab[sy,sx]>=0: continue
            q=deque([(sy,sx)]); lab[sy,sx]=len(sizes); c=0
            while q:
                y,x=q.popleft(); c+=1
                for ny,nx in ((y+1,x),(y-1,x),(y,x+1),(y,x-1)):
                    if 0<=ny<N and 0<=nx<N and not occ[ny,nx] and lab[ny,nx]<0:
                        lab[ny,nx]=len(sizes); q.append((ny,nx))
            sizes.append(c)
    return lab,sizes
for f in sorted(glob.glob('*.tmj')):
    d,L=load(f)
    obs=L['obstacles']
    pts=[(o['x']+o['width']/2,o['y']+o['height']/2,o['type']) for o in L.get('gates',[])+L.get('altars',[])]
    print('==',f)
    for scale in (1,2.5,5,6):
        r=30*scale
        occ=raster(obs,r); lab,sizes=comps(occ)
        free=(~occ).sum()
        big=int(np.argmax(sizes)) if sizes else -1
        frac=max(sizes)/free if free else 0
        unreach=[]
        for x,y,t in pts:
            gx,gy=int(x//G),int(y//G)
            l=lab[gy,gx]
            if l!=big: unreach.append(t)
        print(f'  scale {scale} r={r}: free {free*G*G/12000**2*100:.0f}% of map, main region {frac*100:.0f}% of free, regions {len(sizes)}, gates/altars cut off: {unreach}')

import matplotlib; matplotlib.use('Agg')
import matplotlib.pyplot as plt
from matplotlib.colors import ListedColormap
files=sorted(glob.glob('*.tmj'))
fig,axs=plt.subplots(1,5,figsize=(25,5.6))
for ax,f in zip(axs,files):
    d,L=load(f); obs=L['obstacles']
    occ1=raster(obs,30); occ5=raster(obs,150)
    lab,sizes=comps(occ5); big=int(np.argmax(sizes))
    img=np.zeros((N,N),int)          # 0 free both
    img[occ1]=1                      # obstacle
    lost=(~occ1)&occ5                # blocked only for big
    img[lost]=2
    trapped=(~occ5)&(lab!=big)
    img[trapped]=3
    cmap=ListedColormap(['#eef2e6','#555555','#f4a261','#e63946'])
    ax.imshow(img,cmap=cmap,vmin=0,vmax=3,interpolation='nearest')
    for o in L.get('gates',[]):
        ax.plot((o['x']+o['width']/2)/G,(o['y']+o['height']/2)/G,'o',color='#3a86ff',ms=7)
    for o in L.get('altars',[]):
        ax.plot((o['x']+o['width']/2)/G,(o['y']+o['height']/2)/G,'*',color='#ffbe0b',ms=11,mec='k')
    ax.set_title(d['properties'][0]['value'],fontsize=13); ax.axis('off')
fig.suptitle('Xám = vật cản · Cam = chỗ người Đại Thừa (scale 5) không lọt qua · Đỏ = vùng bị cô lập với người Đại Thừa · Xanh = cổng · Sao = Đài Tiên Khí',fontsize=13)
plt.tight_layout(); plt.savefig('/mnt/user-data/outputs/map_check_scale5.png',dpi=80)
print('saved')
