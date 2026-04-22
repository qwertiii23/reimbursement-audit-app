import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
from matplotlib.patches import FancyBboxPatch, Rectangle, Polygon, Circle
from matplotlib.lines import Line2D
import matplotlib.patheffects as pe
import numpy as np

FONT_TITLE = 'Helvetica Neue'
FONT_CODE = 'Menlo'
FONT_CN = 'PingFang HK'

layers_config = [
    {
        "name_cn": "接入层", "name_en": "ACCESS LAYER",
        "bg": "#EBF4F6", "accent": "#3D8A9E", "text": "#1E5565",
        "items": ["Load Balancer", "DNS", "CDN", "API Endpoint"],
        "ratio": 0.115, "num": "L1"
    },
    {
        "name_cn": "通用网关", "name_en": "GATEWAY LAYER",
        "bg": "#EBECF4", "accent": "#4E5890", "text": "#252D55",
        "items": ["API Gateway", "Auth Service", "Rate Limiter", "Route Config"],
        "ratio": 0.125, "num": "L2"
    },
    {
        "name_cn": "应用层", "name_en": "APPLICATION LAYER",
        "bg": "#FBF4EA", "accent": "#B87828", "text": "#6A4818",
        "groups": [
            ("Application1", ["Func1", "Func2", "Func3"]),
            ("Application2", ["Func4", "Func5"]),
            ("Application3", ["Func6", "Func7", "Func8"]),
        ],
        "ratio": 0.195, "grouped": True, "num": "L3"
    },
    {
        "name_cn": "领域层", "name_en": "DOMAIN LAYER",
        "bg": "#F2EAF5", "accent": "#6E4090", "text": "#381E55",
        "groups": [
            ("Service1", ["Model1", "Repo1"]),
            ("Service2", ["Model2", "Repo2"]),
            ("Service3", ["Model3", "Repo3"]),
        ],
        "ratio": 0.215, "grouped": True, "num": "L4"
    },
    {
        "name_cn": "基建层", "name_en": "INFRASTRUCTURE LAYER",
        "bg": "#EBF1F4", "accent": "#3E6582", "text": "#1E3848",
        "items": ["Message Queue", "Cache Cluster", "Config Center", "Log System", "Monitor"],
        "ratio": 0.15, "num": "L5"
    },
    {
        "name_cn": "数据层", "name_en": "DATA LAYER",
        "bg": "#EBF3ED", "accent": "#3E7550", "text": "#1E4028",
        "items": ["MySQL Cluster", "Redis Cluster", "ES Cluster", "Object Storage"],
        "ratio": 0.14, "num": "L6"
    }
]

fig_w = 19
fig_h = 14
fig, ax = plt.subplots(figsize=(fig_w, fig_h))
ax.set_xlim(0, fig_w)
ax.set_ylim(0, fig_h)
ax.set_aspect('equal')
ax.axis('off')

fig.patch.set_facecolor('#F8F9FA')
ax.set_facecolor('#F8F9FA')

margin_x = 0.7
margin_y_top = 1.15
margin_y_bot = 0.6
draw_x = margin_x + 0.45
draw_y_bot = margin_y_bot
draw_w = fig_w - 2 * margin_x - 0.9
draw_h = fig_h - margin_y_top - margin_y_bot

def hex_to_rgb(h):
    h = h.lstrip('#')
    return tuple(int(h[i:i+2], 16)/255 for i in (0, 2, 4))

def draw_layer(ax, x, y, w, h, config, idx):
    bg = config["bg"]
    accent = config["accent"]
    text_color = config["text"]

    shadow = FancyBboxPatch((x+0.03, y-0.03), w, h,
                             boxstyle="round,pad=0.02,rounding_size=0.14",
                             facecolor='#000000', edgecolor='none',
                             alpha=0.04, zorder=0)
    ax.add_patch(shadow)

    box = FancyBboxPatch((x, y), w, h,
                          boxstyle="round,pad=0.02,rounding_size=0.14",
                          facecolor=bg, edgecolor=accent,
                          linewidth=1.1, alpha=0.97, zorder=1)
    ax.add_patch(box)

    num = config.get("num", "")
    circle_r = 0.18
    cx = x + 0.22
    cy = y + h - 0.35
    circle = Circle((cx, cy), circle_r,
                     facecolor=accent, edgecolor='none',
                     alpha=0.12, zorder=5)
    ax.add_patch(circle)

    ax.text(cx, cy, num,
            fontproperties=FONT_CODE, fontsize=7, fontweight='bold',
            color=accent, va='center', ha='center', zorder=10, alpha=0.8)

    label_x = x + 0.52
    label_y = y + h - 0.30

    ax.text(label_x, label_y, config["name_cn"],
            fontproperties=FONT_CN, fontsize=11.5, fontweight='bold',
            color=accent, va='top', ha='left', zorder=10)

    en_off = len(config["name_cn"]) * 0.055 + 0.18
    ax.text(label_x + en_off, label_y + 0.035, config["name_en"],
            fontproperties=FONT_CODE, fontsize=7,
            color=text_color, va='top', ha='left', alpha=0.65, zorder=10)

    sep_x = x + 0.12
    sep_line = Line2D([sep_x, x+w-0.12], [y+h-0.52, y+h-0.52],
                       color=accent, linewidth=0.35, alpha=0.25, zorder=2,
                       solid_capstyle='round')
    ax.add_line(sep_line)

    if config.get("grouped"):
        groups = config["groups"]
        n_g = len(groups)
        gap_g = 0.25
        group_pad_x = 0.32
        avail_w = w - 2 * group_pad_x - (n_g - 1) * gap_g
        gw = avail_w / n_g
        gy = y + 0.16
        gh = h - 0.72

        for gi, (gname, subs) in enumerate(groups):
            gx = x + group_pad_x + gi * (gw + gap_g)

            gbox = FancyBboxPatch((gx, gy), gw, gh,
                                   boxstyle="round,pad=0.01,rounding_size=0.09",
                                   facecolor='#FFFFFF', edgecolor=accent,
                                   linewidth=0.5, linestyle=(0, (3, 3)),
                                   alpha=0.92, zorder=2)
            ax.add_patch(gbox)

            ax.text(gx + gw/2, gy + gh - 0.24, gname,
                    fontproperties=FONT_CODE, fontsize=8.5, fontweight='bold',
                    color=accent, va='top', ha='center', zorder=10)

            inner_sep = Line2D([gx+0.08, gx+gw-0.08], [gy+gh-0.42, gy+gh-0.42],
                                color=accent, linewidth=0.25, alpha=0.18, zorder=2)
            ax.add_line(inner_sep)

            n_s = len(subs)
            sp = 0.07
            sub_pad = 0.10
            sw_val = (gw - 2*sub_pad - (n_s-1)*sp) / n_s
            sh_val = 0.30
            sy_base = gy + 0.10

            for si, sn in enumerate(subs):
                sx = gx + sub_pad + si*(sw_val+sp)
                al = 0.08 + si * 0.06
                sbox = FancyBboxPatch((sx, sy_base), sw_val, sh_val,
                                       boxstyle="round,pad=0.004,rounding_size=0.05",
                                       facecolor=accent, edgecolor='none',
                                       alpha=al, zorder=3)
                ax.add_patch(sbox)

                ax.text(sx + sw_val/2, sy_base + sh_val/2 - 0.025, sn,
                        fontproperties=FONT_CODE, fontsize=6.8,
                        color=text_color, va='center', ha='center', zorder=10)
    else:
        items = config["items"]
        n_i = len(items)
        gap_i = 0.16
        item_pad_x = 0.32
        iw = (w - 2*item_pad_x - (n_i-1)*gap_i) / n_i
        ih = 0.36
        iy = y + 0.16

        for ii, it in enumerate(items):
            ix = x + item_pad_x + ii*(iw + gap_i)
            al = 0.06 + ii * 0.035
            ibox = FancyBboxPatch((ix, iy), iw, ih,
                                   boxstyle="round,pad=0.004,rounding_size=0.06",
                                   facecolor=accent, edgecolor=accent,
                                   linewidth=0.4, alpha=al, zorder=3)
            ax.add_patch(ibox)

            ax.text(ix + iw/2, iy + ih/2 - 0.025, it,
                    fontproperties=FONT_CODE, fontsize=7,
                    color=text_color, va='center', ha='center', zorder=10)

current_y = draw_y_bot + draw_h
layer_infos = []

for idx, lc in enumerate(layers_config):
    lh = draw_h * lc["ratio"]
    current_y -= lh
    draw_layer(ax, draw_x, current_y, draw_w, lh, lc, idx)
    layer_infos.append({"y_top": current_y + lh, "y_bot": current_y, "config": lc})

for i in range(len(layer_infos) - 1):
    curr = layer_infos[i]
    n_arrows = 5 if curr["config"].get("grouped") else len(curr["config"].get("items", []))
    spacing = draw_w / (n_arrows + 1)

    y_start = curr["y_bot"]
    y_end = y_start - 0.22

    for ai in range(n_arrows):
        ax_pos = draw_x + spacing * (ai + 1)
        ax.annotate('', xy=(ax_pos, y_end), xytext=(ax_pos, y_start),
                    arrowprops=dict(arrowstyle='-|>',
                                    color='#A8B0BC',
                                    lw=0.6,
                                    mutation_scale=5.5,
                                    connectionstyle='arc3,rad=0'),
                    zorder=0)

title_y = draw_y_bot + draw_h + 0.60
ax.text(fig_w/2, title_y, 'SYSTEM ARCHITECTURE',
        fontproperties=FONT_TITLE, fontsize=28, fontweight='light',
        color='#181B21', va='bottom', ha='center', zorder=20)

ax.text(fig_w/2, title_y - 0.42, 'LAYERED ARCHITECTURE DESIGN',
        fontproperties=FONT_CODE, fontsize=7.5,
        color='#9098A4', va='top', ha='center',
        alpha=0.7, zorder=20,
        path_effects=[pe.withStroke(linewidth=0.4, foreground='#F8F9FA')])

dot_left = margin_x + 0.08
dot_right = fig_w - margin_x - 0.08
dot_y = title_y - 0.20
for dx in [dot_left, dot_right]:
    dot = Circle((dx, dot_y), 0.035,
                  facecolor='#C0C8D4', edgecolor='none',
                  alpha=0.5, zorder=20)
    ax.add_patch(dot)

side_label_x = margin_x + 0.08
for i, li in enumerate(layer_infos):
    mid_y = (li["y_top"] + li["y_bot"]) / 2
    accent = li["config"]["accent"]

    tick = Line2D([side_label_x - 0.06, side_label_x], [mid_y, mid_y],
                   color=accent, linewidth=0.8, alpha=0.35, zorder=5,
                   solid_capstyle='butt')
    ax.add_line(tick)

for spine in ['top','right','left','bottom']:
    ax.spines[spine].set_visible(False)

plt.tight_layout(pad=0.5)
output_path = '/Users/lijiaqi/reimbursement-audit/system_architecture.png'
fig.savefig(output_path, dpi=220, bbox_inches='tight',
            facecolor=fig.get_facecolor(), edgecolor='none',
            pad_inches=0.2)
plt.close()
print(f"Refined architecture diagram saved: {output_path}")
