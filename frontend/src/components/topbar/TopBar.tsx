import { createSignal } from "solid-js";
import "./TopBar.css"

const TopBar = () => {
    const [time, setTime] = createSignal(0);
    const [word, setWord] = createSignal("");

    return (
        <div class="TopBar">
            <span class="TopBarLeft">
                {time()}
            </span>
            <span>
                {word()}
            </span>
        </div>
    );
}

export default TopBar;