import { Accessor, Setter, createSignal } from "solid-js";
import { DOMAIN } from "../../App";
import { getLocalPlayer } from "../home/Home.service";
import { Payload, JOIN_CODE, PlayerMsg, Player } from "./messages";

type RoomEvent = (payload: Payload) => void;

export interface RoomConn {
    send: (payload: any) => void;
    subscribe: (code: number, event: RoomEvent) => void;
    players: Accessor<Player[]>
}

export const useRoomConnection = (roomCode: string): RoomConn => {
    let name = getLocalPlayer();
    let socket = new WebSocket(`ws://${DOMAIN}/rooms/join?code=${roomCode}&name=${name}`);
    socket.addEventListener("message", e => console.log(e.data));

    const [players, setPlayers] = createSignal<Player[]>([]);
    socket.addEventListener("message", e => {
        const payload = JSON.parse(e.data) as Payload;
        if (payload.code === JOIN_CODE) {
            const msg = payload.msg as PlayerMsg;
            const newPlayers = [...players()];
            newPlayers[msg.playerIndex] = msg.player;
            setPlayers(newPlayers);
        }
    });

    return {
        send: (payload) => {
            socket.send(JSON.stringify(payload));
        },
        subscribe: (messageCode, event) => {
            socket.addEventListener("message", (e) => {
                const payload = JSON.parse(e.data) as Payload;
                if (payload.code === messageCode) {
                    event(payload);
                }
            });
        },
        players
    }
}