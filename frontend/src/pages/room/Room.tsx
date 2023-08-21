import ScoreBoard from "../../components/scoreboard/ScoreBoard";
import Canvas from "../../components/canvas/Canvas";
import Chat from "../../components/chat/Chat";
import "./Room.css"
import { useParams } from "@solidjs/router";
import TopBar from "../../components/topbar/TopBar";

const Room = () => {
    const { code } = useParams();

    return (
        <div class="Base">
           <div>
            <TopBar />
                <div class="Wrapper">
                    <ScoreBoard />
                    <Canvas />
                    <Chat />
                </div>
           </div>
        </div>
    );
}

export default Room;