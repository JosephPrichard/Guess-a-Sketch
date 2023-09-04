import { createSignal } from "solid-js";
import "./Home.css"
import { createRoom, getLocalPlayer, setLocalPlayer } from "./Home.service";
import { useNavigate } from "@solidjs/router";

const Home = () => {
    const [name, setName] = createSignal(getLocalPlayer());
    const navigate = useNavigate();

    return (
        <div class="Panel CenterPanel">
            <h1 class="HomeTitle">
                Guess a Sketch!
            </h1>
            {/* <Avatar color={0} eyes={0} mouth={0} hat={0} modify /> */}
            <input
                placeholder="Enter your name"
                class="Input"
                value={name()}
                maxLength={15}
                onChange={e => {
                    const newName = e.target.value;
                    setName(newName);
                    setLocalPlayer(newName);
                }}
            />
            <div class="Space"/>
            <button 
                class="Button"
                onClick={async () => {
                    const [value, error] = await createRoom();
                    if (!error) {    
                        navigate(`/rooms/${value}`);
                    } else {
                        console.log(value);
                    }
                }}
            >
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