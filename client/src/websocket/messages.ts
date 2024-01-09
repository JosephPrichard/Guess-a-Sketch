/*
 * Copyright (c) Joseph Prichard 2023
 */

export const [
    START_CODE,
    TEXT_CODE,
    DRAW_CODE,
    CHAT_CODE,
    FINISH_CODE,
    BEGIN_CODE,
    JOIN_CODE,
    LEAVE_CODE,
    TIMEOUT_CODE,
    SAVE_CODE,
    STATE_CODE
] = Array.from(Array(11).keys()).map(i => i+1);

export interface Payload<T = any> {
    code: number;
    msg: T;
}

export interface Player {
    id: string;
    name: string;
}

export interface PlayerMsg {
    player: Player;
    playerIndex: number;
}