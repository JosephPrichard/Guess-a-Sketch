import "./Rooms.css";
import { createSignal, For, onMount, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { DOMAIN } from "../App";

export type RoomType = string;

async function getRooms(): Promise<[RoomType[], string, boolean]> {
    try {
        const resp = await fetch(`http://${DOMAIN}/rooms`, { method: "GET", mode: "cors" });
        const json = await resp.json();
        if (resp.ok) {
            return [json as RoomType[], "", false];
        } else {
            return [[], json["ErrorDesc"] as string, true];
        }
    } catch(ex) {
        return [[], "Failed to get the rooms", true];
    }
}

const Rooms = () => {
    const [rooms, setRooms] = createSignal<RoomType[]>([]);

    const navigate = useNavigate();

    onMount(async () => {
        const [rooms, msg, error] = await getRooms();
        if (!error) {
            setRooms(rooms);
        } else {
            console.log(msg);
        }
    });

    return (
        <div class="PublicRooms">
            <h3 class="HomeSubTitle">
                Public Rooms
            </h3>
            <Show
                when={rooms().length > 0}
                fallback={
                    <div>
                        <div class="Space"/>
                        No public rooms to join!
                    </div>
                }
            >
                <For each={rooms()}>
                    {(room) => (
                        <div class="PublicRoom">
                            <span class="RoomCode">
                                {room}
                            </span>
                            <span class="RoomButtonWrapper">
                                <button
                                    class="Button GreenButton RoomButton"
                                    onClick={async () => {
                                        navigate(`/rooms/${room}`);
                                    }}
                                >
                                    Join
                                </button>
                            </span>
                        </div>
                    )}
                </For>
            </Show>
        </div>
    );
}

export default Rooms;