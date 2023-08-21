import { For } from "solid-js";
import "./ScoreBoard.css";
import { createStore } from "solid-js/store";

interface Score {
    player: string;
    points: number;
}


const ScoreBoard = () => {
    const [scoreBoard, setScoreBoard] = createStore<Score[]>([]);

    return (
        <div class="ScoreBoard">
            <For each={scoreBoard}>
                {(score) => <Score {...score}/>}
            </For>
        </div>
    );
}

const Score = (score: Score) => {
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