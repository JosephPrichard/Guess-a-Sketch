import { createSignal } from "solid-js";
import "./TopBar.css"
import { RoomProps } from "../../pages/room/Room";
import { FINISH_CODE } from "../../pages/room/messages";

const TopBar = ({ room }: RoomProps) => {
    const [time, setTime] = createSignal(0);
    const [word, setWord] = createSignal("Word");
    const [round, setRound] = createSignal(1);
    const [totalRounds, setTotalRound] = createSignal(5);

    room.subscribe(FINISH_CODE, (payload) => {

    });

    return (
        <div class="Panel TopBar">
            <span class="TopElement TopBarLeft">
                {time()}
            </span>
            <span class="TopElement">
                {word()}
            </span>
            <span class="TopElement TopBarRight">
                {round()}/{totalRounds()}
            </span>
        </div>
    );
}

export default TopBar;