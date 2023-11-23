/*
 * Copyright (c) Joseph Prichard 2023
 */

import { For, createSignal } from "solid-js";
import "./ScoreBoard.css";
import { createMutable } from "solid-js/store";
import { RoomProps } from "../pages/Room";

type ScoreBoard = {[key:string]:number};

const ScoreBoard = ({ room }: RoomProps) => {
    const [players, setPlayers] = createSignal<string[]>([]);
    const scoreBoard = createMutable<ScoreBoard>({});

    return (
        <div class="ScoreBoard">
            <div class="TopTitle">
                <span class="ScoreBoardLeft">
                    Scoreboard
                </span>
            </div>
            <div class="ScorePlayers">
                <For each={players()}>
                    {(player) =>  (
                        <div class="Score">
                            <span>
                                {player}
                            </span>
                            <span>
                                {scoreBoard[player]}
                            </span>
                        </div>
                    )}
                </For>
            </div>
        </div>
    );
}

export default ScoreBoard;