/*
 * Copyright (c) Joseph Prichard 2023
 */

import "./Rooms.css";
import { createSignal, For, onMount, Show, useContext } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { BACKEND_URL, TempMsgContext } from "../App";

export type RoomData = string;

async function getRooms(): Promise<[RoomData[], string, boolean]> {
    try {
        const resp = await fetch(`${BACKEND_URL}/rooms`, { method: "GET", mode: "cors" });
        const json = await resp.json();
        if (resp.ok) {
            return [json, "", false];
        } else {
            return [[], json["errorDesc"], true];
        }
    } catch(ex) {
        return [[], "Failed to get the rooms", true];
    }
}

const Rooms = () => {
    const [rooms, setRooms] = createSignal<RoomData[]>([]);

    const navigate = useNavigate();
    const tempMsg = useContext(TempMsgContext);

    onMount(async () => {
        const [rooms, msg, error] = await getRooms();
        if (!error) {
            setRooms(rooms);
        } else {
            tempMsg.addMsg(msg);
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