import "./Room.css"
import ScoreBoard from "../components/ScoreBoard";
import Canvas from "../components/Canvas";
import Chat from "../components/Chat";
import { useParams } from "@solidjs/router";
import TopBar from "../components/TopBar";
import { RoomConn, useRoomConnection } from "./useRoomListener";
import { START_CODE } from "./messages";
import { Match, Show, Switch } from "solid-js";

export interface RoomProps {
    room: RoomConn
}

const Room = () => {
    const { code } = useParams();
    const room = useRoomConnection(code);

    room.subscribe(START_CODE, (payload) => {

    });

    const showModal = () => room.status() !== "opened" && room.status() !== "unopened";

    return (
       <div>
           <Show when={showModal()}>
               <div class="Bg"/>
               <div class="StatusInfo">
                   <Switch>
                       <Match when={room.status() === "closed"}>
                           Disconnected from server.
                       </Match>
                       <Match when={room.status() === "error"}>
                           Unexpected error has occurred.
                       </Match>
                       <Match when={room.status() === "noexist"}>
                           Unable to connect to a room with that code.
                       </Match>
                   </Switch>
               </div>
           </Show>
           <div class="Base">
               <TopBar room={room} code={code}/>
               <div class="Wrapper">
                   <div class="Middle">
                       <Canvas room={room}/>
                   </div>
                   <div class="Sidebar">
                       <ScoreBoard room={room}/>
                       <Chat room={room}/>
                   </div>
               </div>
           </div>
       </div>
    );
}

export default Room;