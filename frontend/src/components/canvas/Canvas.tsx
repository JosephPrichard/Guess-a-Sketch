import "./Canvas.css";
import { Index, createSignal, onMount } from "solid-js";

const COLORS = [
    "white", "grey", "red", "orange", "yellow", "lime", 
    "darkgreen", "cyan", "blue", "purple", "pink", "brown"
];
const SIZES = [4, 6, 8, 12, 16];

const Canvas = () => {
    const [canvasRef, setCanvasRef] = createSignal<HTMLCanvasElement | null>(null);
    const [colorIndex, setColorIndex] = createSignal(0);
    const [sizeIndex, setSizeIndex] = createSignal(0);
    const [isDrawing, setIsDrawing] = createSignal(false);

    const getCanvas = () => {
        const canvas = canvasRef()!;
        const ctx = canvas.getContext("2d")!;
        return { canvas, ctx }
    };

    const drawCircle = (e: MouseEvent) => {
        const { canvas, ctx } = getCanvas();
        const rect = canvas.getBoundingClientRect();
        // calcualte the scale and location to draw the circle
        const scaleX = canvas.width / rect.width;
        const scaleY = canvas.height / rect.height;
        const x = scaleX * (e.clientX - rect.left);
        const y = scaleY * (e.clientY - rect.top);
        // draw the colored circle at the location on the canvas
        ctx.fillStyle = COLORS[colorIndex()];
        ctx.beginPath();
        ctx.arc(x, y, SIZES[sizeIndex()], 0, 2 * Math.PI);
        ctx.fill();
        console.log("Draw circle", x, y, ctx.fillStyle)
    };

    onMount(() => {
        const { canvas, ctx } = getCanvas();
        ctx.fillStyle = "rgb(235, 235, 235)";
        ctx.fillRect(0, 0, canvas.width, canvas.height);
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
                        drawCircle(e);
                    }
                }}
                onmousedown={(e) => {
                    drawCircle(e);
                    setIsDrawing(true);
                }}
                onmouseup={() => setIsDrawing(false)}
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