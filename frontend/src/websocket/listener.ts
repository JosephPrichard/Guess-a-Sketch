import { Accessor, createSignal, onCleanup, onMount } from "solid-js";
import { WS_URL } from "../App";
import { Payload, JOIN_CODE, PlayerMsg, Player } from "./messages";
type RoomEvent = (payload: Payload) => void;
type RoomStatus = "opened" | "closed" | "error" | "unopened" | "noexist";

export interface RoomConn {
    send: <T extends Payload>(payload: T) => void;
    subscribe: (code: number, event: RoomEvent) => void;
    players: Accessor<Player[]>;
    status: Accessor<RoomStatus>;
}

function getLocalPlayer() {
    return localStorage.getItem("player") ?? "";
}

export const useRoomConnection = (roomCode: string): RoomConn => {
    let name = getLocalPlayer();

    const [socket, setSocket] = createSignal<WebSocket>();
    const [status, setStatus] = createSignal<RoomStatus>("unopened");
    const [players, setPlayers] = createSignal<Player[]>([]);
    const handlerMap = new Map<number, RoomEvent[]>();

    onMount(() => {
        const token = localStorage.getItem("session-token") ?? "";
        setSocket(new WebSocket(`${WS_URL}/rooms/join?code=${roomCode}&name=${name}&token=${token}`));

        const sock = socket();
        sock?.addEventListener("message", e => {
            console.log("Receiving message", e.data);

            const payload = JSON.parse(e.data) as Payload;
            const handlers = handlerMap.get(payload.code);
            if (handlers) {
                for (const handler of handlers) {
                    handler(payload);
                }
            }
        });

        sock?.addEventListener("close", () => {
            if (status() !== "opened") {
                setStatus("noexist");
            } else {
                setStatus("closed");
            }
        });
        sock?.addEventListener("open", () => setStatus("opened"))

        onCleanup(() => socket()?.close());
    });

    const send = (payload: Payload) => {
        const m = JSON.stringify(payload);
        console.log("Sending message", m);
        const sock = socket();
        sock?.send(m);
    };

    const subscribe = (code: number, event: RoomEvent) => {
        let handlers = handlerMap.get(code);
        if (!handlers) {
            handlers = [];
        }
        handlers.push(event);
        handlerMap.set(code, handlers);
    };

    subscribe(JOIN_CODE, (payload) => {
        const msg = payload.msg as PlayerMsg;
        const newPlayers = [...players()];
        newPlayers[msg.playerIndex] = msg.player;
        setPlayers(newPlayers);
    });

    return {send, subscribe, players, status};
}