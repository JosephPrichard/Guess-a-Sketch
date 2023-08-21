import "./Canvas.css";
import { Index, createSignal, onMount } from "solid-js";

const COLORS = [
    "white", "grey", "red", "orange", "yellow", "lime", 
    "darkgreen", "cyan", "blue", "purple", "pink", "brown"
];
const SIZES = [2, 4, 8, 16, 32];

const Canvas = () => {
    const [canvasRef, setCanvasRef] = createSignal<HTMLCanvasElement | null>(null);
    const [color, setColor] = createSignal(0);
    const [size, setSize] = createSignal(0);

    const getCanvas = () => {
        const canvas = canvasRef()!;
        const ctx = canvas.getContext("2d")!;
        return { canvas, ctx }
    };

    onMount(() => {
        const { canvas, ctx } = getCanvas();
        ctx.fillStyle = "rgb(235, 235, 235)";
        ctx.fillRect(0, 0, canvas.width, canvas.height);
    });

    return (
        <div>
            <canvas ref={setCanvasRef} class="Canvas" />
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
        </div>
    );
}

export default Canvas;