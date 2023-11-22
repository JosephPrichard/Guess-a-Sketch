import { createSignal, Show } from "solid-js";
import "./Home.css"
import { useNavigate } from "@solidjs/router";
import CreateRoom from "../components/CreateRoom";
import Rooms from "../components/Rooms";
import Login from "../components/Login";

function setLocalPlayer(player: string) {
    localStorage.setItem("player", player);
}

function getLocalPlayer() {
    return localStorage.getItem("player") ?? "";
}

const Home = () => {
    const [name, setName] = createSignal(getLocalPlayer());
    const [code, setCode] = createSignal("");
    const [showCreate, setShowCreate] = createSignal(false);

    const navigate = useNavigate();

    return (
        <div>
            <Show when={showCreate()}>
                <CreateRoom onClose={() => setShowCreate(false)}/>
            </Show>
            <div class="Home">
                <h3 class="HomeSubTitle">
                    Welcome to
                </h3>
                <h1 class="HomeTitle">
                    Guess a Sketch!
                </h1>
                <div class="HomeBody">
                    Guess a Sketch is a free, online drawing & guessing game similar to Pictionary that can be played on any browser!
                    Play with your friends or with anyone around the world. Login to save your stats and compete with your friends.
                </div>
                <div class="Space"/>
                <Login/>
                <div class="Space"/>
                <div class="Group">
                    <input
                        placeholder="Enter your name"
                        class="Input"
                        value={name()}
                        maxLength={15}
                        onChange={e => {
                            const newName = e.target.value;
                            setName(newName);
                        }}
                    />
                    <div class="Space"/>
                    <button class="Button" onClick={() => setLocalPlayer(name())}>
                        Change Name
                    </button>
                </div>
                <div class="Space"/>
                <div class="Group">
                    <input
                        placeholder="Enter Room Code"
                        class="Input"
                        value={code()}
                        maxLength={15}
                        onChange={e => {
                            const newCode = e.target.value;
                            setCode(newCode);
                        }}
                    />
                    <div class="Space"/>
                    <button class="Button" onClick={async () => navigate(`/rooms/${code()}`)}>
                        Join Room
                    </button>
                </div>
                <div class="Space"/>
                <div class="Group">
                    <button class="Button GreenButton" onClick={() => setShowCreate(true)}>
                        Create Room
                    </button>
                    <div class="Space"/>
                    <button class="Button">
                        Quick Play
                    </button>
                </div>
                <Rooms/>
            </div>
        </div>
    );
}

export default Home;