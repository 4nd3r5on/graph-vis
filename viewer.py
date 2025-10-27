#!/usr/bin/env python3
"""
viewer.py — simple DAG visualizer using pygame.

Usage:
    python viewer.py input.json out.json

Controls:
    Left mouse drag on empty space -> pan (translate) the canvas
    Shift + Left drag on a node -> move that node (updates positions in-memory)
    S key -> save current positions to out.json (overwrites)
    Esc or window close -> exit
"""

import sys
import json
import math
import pygame
from pathlib import Path
from typing import Tuple, Dict, Any

NODE_RADIUS = 28
FONT_SIZE = 16
EDGE_LABEL_MARGIN = 6
BACKGROUND = (30, 30, 30)
NODE_COLOR = (50, 130, 230)
NODE_BORDER = (220, 220, 220)
EDGE_COLOR = (200, 200, 200)
EDGE_LABEL_BG = (20, 20, 20)
TEXT_COLOR = (245, 245, 245)
FPS = 60

def load_input(input_path: Path) -> Dict[str, Any]:
    with input_path.open('r', encoding='utf-8') as f:
        return json.load(f)

def load_out(out_path: Path) -> Dict[int, Dict[str, float]]:
    if not out_path.exists():
        return {}
    with out_path.open('r', encoding='utf-8') as f:
        arr = json.load(f)
    return {int(x['id']): {'x': float(x['pos']['x']), 'y': float(x['pos']['y'])} for x in arr}

def save_out(out_path: Path, node_positions: Dict[int, Dict[str, float]]):
    arr = [{'id': nid, 'pos': {'x': pos['x'], 'y': pos['y']}} for nid, pos in node_positions.items()]
    with out_path.open('w', encoding='utf-8') as f:
        json.dump(arr, f, indent=2)
    print(f"Saved {len(arr)} nodes to {out_path}")

def to_screen(world_pos: Tuple[float, float], offset: Tuple[float, float]) -> Tuple[int, int]:
    x = int(world_pos[0] + offset[0])
    y = int(world_pos[1] + offset[1])
    return x, y

def distance(a: Tuple[int,int], b: Tuple[int,int]) -> float:
    dx = a[0]-b[0]; dy = a[1]-b[1]; return math.hypot(dx, dy)

def draw_rounded_rect(surface, rect, color, radius=6):
    pygame.draw.rect(surface, color, rect, border_radius=radius)

def main():
    if len(sys.argv) < 3:
        print("Usage: python viewer.py input.json out.json")
        sys.exit(1)
    input_path = Path(sys.argv[1])
    out_path = Path(sys.argv[2])

    data = load_input(input_path)
    positions = load_out(out_path)

    # Ensure positions exist for every node (fallback to grid if missing)
    nodes = data.get('nodes', {})
    # convert keys (strings) to ints
    node_ids = [int(k) for k in nodes.keys()]
    for i, nid in enumerate(sorted(node_ids)):
        if nid not in positions:
            # fallback grid placement
            positions[nid] = {'x': (i % 10) * 240 + 200, 'y': (i // 10) * 160 + 120}

    pygame.init()
    screen = pygame.display.set_mode((1200, 800))
    pygame.display.set_caption("DAG Viewer")
    clock = pygame.time.Clock()
    font = pygame.font.SysFont(None, FONT_SIZE)
    edge_font = pygame.font.SysFont(None, max(12, FONT_SIZE-2))

    offset_x, offset_y = -100, -50  # initial translation of world -> screen
    panning = False
    pan_last = (0,0)

    dragging_node = None  # node id being dragged while holding Shift
    drag_node_offset = (0,0)

    running = True
    while running:
        for ev in pygame.event.get():
            if ev.type == pygame.QUIT:
                running = False
            elif ev.type == pygame.KEYDOWN:
                if ev.key == pygame.K_ESCAPE:
                    running = False
                elif ev.key == pygame.K_s:
                    save_out(out_path, positions)
            elif ev.type == pygame.MOUSEBUTTONDOWN:
                if ev.button == 1:  # left mouse down
                    mx, my = ev.pos
                    # check if Shift pressed and clicked on node -> start node drag
                    mods = pygame.key.get_mods()
                    clicked_node = None
                    for nid, pos in positions.items():
                        sx, sy = to_screen((pos['x'], pos['y']), (offset_x, offset_y))
                        if distance((sx,my),(sx,my)) is None:
                            pass
                        if distance((mx,my),(sx,sy)) <= NODE_RADIUS:
                            clicked_node = nid
                            break
                    if mods & pygame.KMOD_SHIFT and clicked_node is not None:
                        dragging_node = clicked_node
                        sx, sy = to_screen((positions[clicked_node]['x'], positions[clicked_node]['y']), (offset_x, offset_y))
                        drag_node_offset = (sx - mx, sy - my)
                    else:
                        # start panning
                        panning = True
                        pan_last = ev.pos
                elif ev.button == 3:  # right mouse down -> also pan
                    panning = True
                    pan_last = ev.pos
            elif ev.type == pygame.MOUSEBUTTONUP:
                if ev.button in (1,3):
                    panning = False
                    dragging_node = None
            elif ev.type == pygame.MOUSEMOTION:
                if panning:
                    mx, my = ev.pos
                    dx = mx - pan_last[0]
                    dy = my - pan_last[1]
                    offset_x += dx
                    offset_y += dy
                    pan_last = (mx, my)
                if dragging_node is not None:
                    mx, my = ev.pos
                    sx = mx + drag_node_offset[0]
                    sy = my + drag_node_offset[1]
                    # update world coordinates inversely by offset
                    positions[dragging_node]['x'] = sx - offset_x
                    positions[dragging_node]['y'] = sy - offset_y

        screen.fill(BACKGROUND)

        # draw edges first
        for pid_str, node in nodes.items():
            pid = int(pid_str)
            ppos = positions.get(pid)
            if not ppos:
                continue
            p_screen = to_screen((ppos['x'], ppos['y']), (offset_x, offset_y))
            children = node.get('children', [])
            for ch in children:
                cid = int(ch['id'])
                cpos = positions.get(cid)
                if not cpos:
                    continue
                c_screen = to_screen((cpos['x'], cpos['y']), (offset_x, offset_y))
                # compute line endpoints that stop at node border (circle)
                dx = c_screen[0] - p_screen[0]
                dy = c_screen[1] - p_screen[1]
                dist = math.hypot(dx, dy) if (dx or dy) else 1.0
                ux = dx / dist; uy = dy / dist
                start = (int(p_screen[0] + ux * NODE_RADIUS), int(p_screen[1] + uy * NODE_RADIUS))
                end = (int(c_screen[0] - ux * NODE_RADIUS), int(c_screen[1] - uy * NODE_RADIUS))
                pygame.draw.line(screen, EDGE_COLOR, start, end, 2)

                # draw simple arrowhead
                ah_size = 10
                left = (end[0] - int(ux * ah_size) - int(uy * ah_size/2),
                        end[1] - int(uy * ah_size) + int(ux * ah_size/2))
                right = (end[0] - int(ux * ah_size) + int(uy * ah_size/2),
                         end[1] - int(uy * ah_size) - int(ux * ah_size/2))
                pygame.draw.polygon(screen, EDGE_COLOR, [end, left, right])

                # edge label at midpoint
                label = ch.get('connLabel') or ''
                if label:
                    mx = (start[0] + end[0]) // 2
                    my = (start[1] + end[1]) // 2
                    text_surf = edge_font.render(label, True, TEXT_COLOR)
                    bg_rect = text_surf.get_rect(center=(mx, my))
                    bg_rect.inflate_ip(6, 4)
                    draw_rounded_rect(screen, bg_rect, EDGE_LABEL_BG)
                    screen.blit(text_surf, text_surf.get_rect(center=(mx,my)))

        # draw nodes on top
        for nid, pos in positions.items():
            sx, sy = to_screen((pos['x'], pos['y']), (offset_x, offset_y))
            pygame.draw.circle(screen, NODE_COLOR, (sx, sy), NODE_RADIUS)
            pygame.draw.circle(screen, NODE_BORDER, (sx, sy), NODE_RADIUS, 2)
            label = str(nodes.get(str(nid), {}).get('label', str(nid)))
            text_surf = font.render(label, True, TEXT_COLOR)
            ts_rect = text_surf.get_rect(center=(sx, sy))
            screen.blit(text_surf, ts_rect)

        # small HUD
        fps_text = font.render(f"FPS: {int(clock.get_fps())}  Nodes: {len(positions)}  Pan: LMB/ RMB drag  Shift+L drag node  S=save", True, (200,200,200))
        screen.blit(fps_text, (8,8))

        pygame.display.flip()
        clock.tick(FPS)

    pygame.quit()

if __name__ == "__main__":
    main()
