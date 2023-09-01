import { For } from "solid-js";
import "./ScoreBoard.css";
import { createStore } from "solid-js/store";
import { RoomProps } from "../../pages/room/Room";

const ScoreBoard = ({ room }: RoomProps) => {
    const [scoreBoard, setScoreBoard] = createStore<ScoreProps[]>([]);

    return (
        <div class="ScoreBoard">
            <For each={scoreBoard}>
                {(score) => <Score {...score}/>}
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
            <span class="BoldText ScorePiece">
                {score.player}
            </span>
            <span class="ScorePiece">
                {score.points}
            </span>
        </div>
    );
}

export default ScoreBoard;