/*
 * Copyright (c) Joseph Prichard 2024
 */

import { Player } from "./messages";
import { ScoreBoard } from "../components/ScoreBoard";
import { ChatMsg } from "../components/Chat";

interface Turn {
    currWord: string;
    currPlayer: Player;
    canvas: string;
}

export interface StateMsg {
    currRound: number;
    players: Player[];
    scoreBoard: ScoreBoard;
    chatLog: ChatMsg[];
    stage: number;
    turn: Turn;
}