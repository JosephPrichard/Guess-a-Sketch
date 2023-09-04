import "./Room.css"
import ScoreBoard from "../../components/scoreboard/ScoreBoard";
import Canvas from "../../components/canvas/Canvas";
import Chat from "../../components/chat/Chat";
import { useParams } from "@solidjs/router";
import TopBar from "../../components/topbar/TopBar";
import { RoomConn, useRoomConnection } from "./useRoomListener";
import { START_CODE } from "./messages";
import Settings from "../../components/settings/Settings";

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
           {/* <div>
                <TopBar room={room} />
                <div class="Wrapper">
                    <Canvas room={room} />
                    <Chat room={room} />
                </div>
                <ScoreBoard room={room} />
           </div> */}
           <Settings room={room} />
        </div>
    );
}

export default Room;