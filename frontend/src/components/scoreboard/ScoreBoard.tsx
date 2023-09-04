import { For, createSignal } from "solid-js";
import "./ScoreBoard.css";
import { createMutable } from "solid-js/store";
import { RoomProps } from "../../pages/room/Room";

type ScoreBoard = {[key:string]:number};

const ScoreBoard = ({ room }: RoomProps) => {
    const [players, setPlayers] = createSignal<string[]>([]);
    const scoreBoard = createMutable<ScoreBoard>({});

    return (
        <div class="Panel ScoreBoard">
            <For each={players()}>
                {(player) => {
                    return (
                        <Score 
                            player={player} 
                            points={scoreBoard[player]}
                        />
                    );
                }}
            </For>
        </div>
    );
}

interface ScoreProps {
    player: string;
    points: number;
}

const Score = (score: ScoreProps) => {
    return (
        <div class="Score">
            <div class="BoldText">
                {score.player}
            </div>
            <div class="ScoreSubtitle">
                Score: {score.points}
            </div>
        </div>
    );
}

export default ScoreBoard;