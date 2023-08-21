import { createSignal } from "solid-js";
import "./Home.css"

const Home = () => {
    const [name, setName] = createSignal("");

    return (
        <div class="CenterPanel">
            <input
                placeholder="Enter your name"
                class="Input"
                value={name()}
                onChange={e => setName(e.target.value)}
            />
            <div class="Space"/>
            <button class="Button">
                Create Room
            </button>
            <div class="Space"/>
            <button class="Button">
                Play Online
            </button>
        </div>
    );
}

export default Home;