import { DOMAIN } from "../../App";
import { getLocalPlayer } from "../home/Home.service";

interface Payload {
    code: number;
}

type RoomEvent = (payload: Payload) => void;

export interface RoomConn {
    send: (payload: any) => void;
    subscribe: (code: number, event: RoomEvent) => void;
}

export const useRoomConnection = (code: string): RoomConn => {
    let player = getLocalPlayer();
    let socket = new WebSocket(`ws://${DOMAIN}/rooms?code=${code}&player=${player}`);

    return {
        send: (payload) => {
            socket.send(JSON.stringify(payload));
        },
        subscribe: (code, event) => {
            socket.addEventListener("message", (e) => {
                const payload = JSON.parse(e.data) as Payload;
                if (payload.code === code) {
                    event(payload);
                }
            });
        }
    }
}