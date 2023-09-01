import "./Room.css"
import ScoreBoard from "../../components/scoreboard/ScoreBoard";
import Canvas from "../../components/canvas/Canvas";
import Chat from "../../components/chat/Chat";
import { useParams } from "@solidjs/router";
import TopBar from "../../components/topbar/TopBar";
import { RoomConn, useRoomConnection } from "./useRoomListener";
import { START_CODE } from "./messages";

export interface RoomProps {
    room: RoomConn
}

const Room = () => {
    const { code } = useParams();
    const room = useRoomConnection(code);

    room.subscribe(START_CODE, (payload) => {

    });

    return (
        <div class="Base">
           <div>
                <TopBar room={room} />
                <div class="Wrapper">
                    <ScoreBoard room={room} />
                    <Canvas room={room} />
                    <Chat room={room} />
                </div>
           </div>
        </div>
    );
}

export default Room;