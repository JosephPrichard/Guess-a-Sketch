/*
 * Copyright (c) Joseph Prichard 2023
 */

import "./Canvas.css";
import { Index, createSignal, onMount } from "solid-js";
import { RoomProps } from "../pages/Room";
import { DRAW_CODE, STATE_CODE } from "../websocket/messages";
import { ChatMsg } from "./Chat";
import { StateMsg } from "../websocket/state";

const COLORS = [
    "black", "white", "grey", "red", "orange", "yellow", "lime", 
    "darkgreen", "cyan", "blue", "purple", "pink", "brown",
];
const RADII = [4, 6, 8, 12, 16, 20];

interface Point {
    x: number;
    y: number;
}

interface Circle {
    color: number;
    radius: number;
    x: number;
    y: number;
    connected?: boolean;
}

type DragMsg = Circle;

function drawCircle(ctx: CanvasRenderingContext2D, radius: number, point: Point) {
    ctx.beginPath();
    ctx.arc(point.x, point.y, radius, 0, 2 * Math.PI);
    ctx.fill();
}

function interpolate(ctx: CanvasRenderingContext2D, radius: number, from: Point, to: Point) {
    const xDiff = to.x - from.x;
    const yDiff = to.y - from.y;
    const segments = Math.sqrt(Math.pow(xDiff, 2) + Math.pow(yDiff, 2)) / radius;
    for (let i = 0; i < segments; i++) {
        const nextPoint = {
            x: from.x + (xDiff / segments) * i,
            y: from.y + (yDiff / segments) * i
        };
        drawCircle(ctx, radius, nextPoint);
    }
}

function at(binary: Uint8Array, i: number) {
    const b = binary.at(i);
    if (b === undefined) {
        throw new Error("Error decoding binary data");
    }
    return b;
}

function circlesFromB64(b64Canvas: string) {
    // decode the binary64 message to a binary array
    const binStrCanvas = atob(b64Canvas);
    const binaryCanvas = new Uint8Array(binStrCanvas.length);
    for (let i = 0; i < binStrCanvas.length; i++) {
        binaryCanvas[i] = binStrCanvas.charCodeAt(i);
    }

    // write the binary array into an in memory representation
    const circles: Circle[] = [];
    for (let i = 0; b64Canvas.length; i += 7) {
        const binaryCircle = binaryCanvas.slice(i, i+7);
        circles.push({
            color: at(binaryCircle,0),
            radius: at(binaryCircle,1),
            x: at(binaryCircle,2) & at(binaryCircle,3) >> 8,
            y: at(binaryCircle,4) & at(binaryCircle,5) >> 8,
            connected: at(binaryCircle,6) > 0
        });
    }
    return circles;
}

const Canvas = ({ room }: RoomProps) => {
    const [canvasRef, setCanvasRef] = createSignal<HTMLCanvasElement | null>(null);
    const [colorIndex, setColorIndex] = createSignal(0);
    const [radiusIndex, setRadiusIndex] = createSignal(0);
    const [isDrawing, setIsDrawing] = createSignal(false);
    const [prevPos, setPrevPos] = createSignal<Point | null>(null);

    const getCanvas = () => {
        const canvas = canvasRef()!;
        const ctx = canvas.getContext("2d")!;
        return { canvas, ctx };
    };

    const drawStroke = (c: Circle, lastPos: Point | null): Point => {
        const { canvas, ctx } = getCanvas();
        ctx.fillStyle = COLORS[c.color];
        const rect = canvas.getBoundingClientRect();

        const scaleX = canvas.width / rect.width;
        const scaleY = canvas.height / rect.height;
        const point = {
            x: scaleX * (c.x - rect.left),
            y: scaleY * (c.y - rect.top)
        };

        const radius = RADII[c.radius];
        drawCircle(ctx, radius, point);

        if (lastPos) {
            interpolate(ctx, radius, point, lastPos);
        }

        return point;
    }

    const onDrawEvent = (c: Circle) => {
        const lastPos = prevPos();
        const point = drawStroke(c, lastPos);
        setPrevPos(point);
    };

    const onDrawMouseEvent = (e: MouseEvent) => {
        const c: Circle = {
            x: e.clientX,
            y: e.clientY,
            color: colorIndex(),
            radius: radiusIndex()
        }
        onDrawEvent(c);
    }

    room.subscribe<StateMsg>(STATE_CODE, (payload) => {
        const b64Canvas = payload.msg.turn.canvas;
        const circles = circlesFromB64(b64Canvas);

        let lastPos: Point | null = null;
        for (const c of circles) {
            drawStroke(c, lastPos);
            lastPos = {x: c.x, y: c.y};
        }
    });

    room.subscribe<DragMsg>(DRAW_CODE, (payload) => {
        const msg = payload.msg;
        if (!msg.connected) {
            setPrevPos(null);
        }
        onDrawEvent(msg);
    });

    return (
        <div class="Canvas">
            <canvas 
                ref={setCanvasRef} 
                width="1200"
                height="1200"
                style={{
                    height: "calc(100vh - 50px - 30px)"
                }}
                onmousemove={(e) => {
                    if (isDrawing()) {
                        onDrawMouseEvent(e);
                    }
                }}
                onmousedown={(e) => {
                    onDrawMouseEvent(e);
                    setIsDrawing(true);
                }}
                onmouseup={() => {
                    setPrevPos(null);
                    setIsDrawing(false);
                }}
                onblur={() => setIsDrawing(false)}
                onmouseleave={() => setIsDrawing(false)}
            />
            <Easel setColor={setColorIndex} setSize={setRadiusIndex} />
        </div>
    );
}

interface EaselProps {
    setColor: (c: number) => void;
    setSize: (s: number) => void;
}

const Easel = ({ setColor, setSize }: EaselProps) => {
    return (
        <div class="Easel">
            <div class="ColorWrapper">
                <Index each={COLORS}>
                    {(color, i) => (
                        <div 
                            class="ColorTile" 
                            style={{ "background-color": color() }} 
                            onclick={() => setColor(i)}
                        />
                    )}
                </Index>
            </div>
        </div>
    );
}

export default Canvas;