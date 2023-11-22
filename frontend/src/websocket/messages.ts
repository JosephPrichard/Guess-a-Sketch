export const [
    OPTIONS_CODE,
    START_CODE,
    TEXT_CODE,
    DRAW_CODE,
    CHAT_CODE,
    FINISH_CODE,
    BEGIN_CODE,
    JOIN_CODE,
    LEAVE_CODE,
    TIMEOUT_CODE
] = Array.from(Array(11).keys());

export interface Payload {
    code: number;
    msg: any;
}

export interface Player {
    id: string;
    name: string;
}

export interface PlayerMsg {
    player: Player;
    playerIndex: number;
}