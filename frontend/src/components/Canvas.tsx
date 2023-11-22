import "./Canvas.css";
import { Index, createSignal, onMount } from "solid-js";
import { RoomProps } from "../pages/Room";
import { DRAW_CODE } from "../websocket/messages";

const COLORS = [
    "black", "white", "grey", "red", "orange", "yellow", "lime", 
    "darkgreen", "cyan", "blue", "purple", "pink", "brown",
];
const SIZES = [4, 6, 8, 12, 16, 20];

interface Point { 
    x: number; 
    y: number;
}

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
        drawCircle(ctx, radius, nextPoint)
    }
}

const Canvas = ({ room }: RoomProps) => {
    const [canvasRef, setCanvasRef] = createSignal<HTMLCanvasElement | null>(null);
    const [colorIndex, setColorIndex] = createSignal(0);
    const [sizeIndex, setSizeIndex] = createSignal(0);
    const [isDrawing, setIsDrawing] = createSignal(false);
    const [prevPos, setPrevPos] = createSignal<Point | null>(null)

    const getCanvas = () => {
        const canvas = canvasRef()!;
        const ctx = canvas.getContext("2d")!;
        return { canvas, ctx };
    };

    const onDrawEvent = (e: MouseEvent) => {
        const { canvas, ctx } = getCanvas();
        ctx.fillStyle = COLORS[colorIndex()];
        const rect = canvas.getBoundingClientRect();

        const scaleX = canvas.width / rect.width;
        const scaleY = canvas.height / rect.height;
        const point = {
            x: scaleX * (e.clientX - rect.left),
            y: scaleY * (e.clientY - rect.top)
        };

        const radius = SIZES[sizeIndex()];
        drawCircle(ctx, radius, point);

        const lastPos = prevPos();
        if (lastPos) {
            interpolate(ctx, radius, point, lastPos);
        }

        setPrevPos(point);
    };

    room.subscribe(DRAW_CODE, (payload) => {

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
                        onDrawEvent(e);
                    }
                }}
                onmousedown={(e) => {
                    onDrawEvent(e);
                    setIsDrawing(true);
                }}
                onmouseup={() => {
                    setPrevPos(null);
                    setIsDrawing(false);
                }}
                onblur={() => setIsDrawing(false)}
                onmouseleave={() => setIsDrawing(false)}
            />
            <Easel setColor={setColorIndex} setSize={setSizeIndex} />
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