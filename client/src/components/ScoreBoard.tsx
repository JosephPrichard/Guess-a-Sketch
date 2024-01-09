/*
 * Copyright (c) Joseph Prichard 2023
 */

import { For, createSignal } from "solid-js";
import "./ScoreBoard.css";
import { createMutable } from "solid-js/store";
import { RoomProps } from "../pages/Room";

export type Score = {points: number}

export type ScoreBoard = {[key:string]:Score};

const ScoreBoard = ({ room }: RoomProps) => {
    const scoreBoard = createMutable<ScoreBoard>({});

    return (
        <div class="ScoreBoard">
            <div class="TopTitle">
                <span class="ScoreBoardLeft">
                    Scoreboard
                </span>
            </div>
            <div class="ScorePlayers">
                <For each={room.players()}>
                    {(player) =>  (
                        <div class="Score">
                            <span>
                                {player.name}
                            </span>
                            <span>
                                {scoreBoard[player.id].points}
                            </span>
                        </div>
                    )}
                </For>
            </div>
        </div>
    );
}

export default ScoreBoard;