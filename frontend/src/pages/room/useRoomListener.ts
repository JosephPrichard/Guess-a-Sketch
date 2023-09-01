const DOMAIN = "localhost:8080";

interface Payload {
    code: number;
}

type RoomEvent = (payload: Payload) => void;

export interface RoomConn {
    send: (payload: any) => void;
    subscribe: (code: number, event: RoomEvent) => void;
}

export const useRoomConnection = (code: string): RoomConn => {
    let socket = new WebSocket(`ws://${DOMAIN}/rooms/${code}`);

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