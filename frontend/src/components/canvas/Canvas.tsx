import "./Canvas.css";
import { Index, createSignal, onMount } from "solid-js";
import { RoomProps } from "../../pages/room/Room";
import { DRAW_CODE } from "../../pages/room/messages";

const COLORS = [
    "white", "grey", "red", "orange", "yellow", "lime", 
    "darkgreen", "cyan", "blue", "purple", "pink", "brown"
];
const SIZES = [4, 6, 8, 12, 16];

interface Point { 
    x: number; 
    y: number;
};

function drawCircle(ctx: CanvasRenderingContext2D, radius: number, point: Point) {
    ctx.beginPath();
    ctx.arc(point.x, point.y, radius, 0, 2 * Math.PI);
    ctx.fill();
};

function interpolate(radius: number, from: Point, to: Point) {
    const xDiff = to.x - from.x;
    const yDiff = to.y - from.y;
    const segments = Math.sqrt(Math.pow(xDiff, 2) + Math.pow(yDiff, 2)) / radius;
    const points: Point[] = [];
    for (let i = 0; i < segments; i++) {
        const nextPoint = {
            x: from.x + xDiff / segments,
            y: from.y + yDiff / segments
        };
        points.push(nextPoint);
    }
    return points;
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
        return { canvas, ctx }
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

        let points: Point[] = [point];

        const radius = SIZES[sizeIndex()];
        const lastPos = prevPos();
        if (lastPos) {
           points.concat(interpolate(radius, point, lastPos));
        }
    
        for (const p of points) {
            drawCircle(ctx, radius, p);
        }

        setPrevPos(point);
    };

    onMount(() => {
        const { canvas, ctx } = getCanvas();
        ctx.fillStyle = "rgb(235, 235, 235)";
        ctx.fillRect(0, 0, canvas.width, canvas.height);
    });

    room.subscribe(DRAW_CODE, (payload) => {

    });

    return (
        <div>
            <canvas 
                ref={setCanvasRef} 
                class="Canvas"
                width="700"
                height="450"
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
        <div class="Paint">
            <div>
                <Index each={COLORS}>
                    {(color, i) => (
                        <div 
                            class="SmallTile" 
                            style={{ "background-color": color() }} 
                            onclick={() => setColor(i)}
                        />
                    )}
                </Index>
            </div>
            <div>
                <Index each={SIZES}>
                    {(size, i) => (
                        <div 
                            class="LargeTile"
                            onclick={() => setSize(i)}
                        />
                    )}
                </Index>
                <div class="LargeTile" />
            </div>
        </div>
    );
}

export default Canvas;