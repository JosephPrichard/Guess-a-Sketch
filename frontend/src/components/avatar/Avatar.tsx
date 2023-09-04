import { createSignal, onMount } from "solid-js";
import "./Avatar.css";

interface AvatarProps {
    color: number;
    eyes: number;
    mouth: number;
    hat: number;
    modify?: boolean;
}

const Avatar = ({ color, eyes, mouth, hat, modify }: AvatarProps) => {
    const [canvasRef, setCanvasRef] = createSignal<HTMLCanvasElement | null>(null);

    const getCanvas = () => {
        const canvas = canvasRef()!;
        const ctx = canvas.getContext("2d")!;
        return { canvas, ctx };
    };

    onMount(() => {
        const { canvas, ctx } = getCanvas();
        ctx.fillStyle = "black";
        ctx.lineWidth = 3;
        ctx.imageSmoothingEnabled = false;
        // draws the player model
        ctx.arc(canvas.width / 2, canvas.width / 3, canvas.width / 5.5, 0, 2 * Math.PI);
        ctx.stroke();
    });

    return (
        <div class="AvatarWrapper">
            <canvas 
                ref={setCanvasRef}
                width="150px"
                height="150px"
            />
        </div>
    );
}

export default Avatar;