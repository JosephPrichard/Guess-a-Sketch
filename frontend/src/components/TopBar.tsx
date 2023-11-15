import { createSignal } from "solid-js";
import "./TopBar.css"
import { RoomProps } from "../pages/Room";
import { FINISH_CODE } from "../room/messages";

type Props = RoomProps & { code: string };

const TopBar = ({ room, code }: Props) => {
    const [time, setTime] = createSignal(0);
    const [word, setWord] = createSignal("Word");
    const [round, setRound] = createSignal(1);
    const [totalRounds, setTotalRound] = createSignal(5);

    room.subscribe(FINISH_CODE, (payload) => {

    });

    return (
        <div class="TopBar">
            <span class="TopBarLeft">
                Round {round()}/{totalRounds()}
            </span>
            <span>
                Guessing {word()}
            </span>
            <span class="TopBarRight">
                {time()} seconds
            </span>
        </div>
    );
}

export default TopBar;