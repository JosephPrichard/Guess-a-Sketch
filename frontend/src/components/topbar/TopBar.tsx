import { createSignal } from "solid-js";
import "./TopBar.css"
import { RoomProps } from "../../pages/room/Room";
import { FINISH_CODE } from "../../pages/room/messages";

const TopBar = ({ room }: RoomProps) => {
    const [time, setTime] = createSignal(0);
    const [word, setWord] = createSignal("");

    room.subscribe(FINISH_CODE, (payload) => {

    });

    return (
        <div class="TopBar">
            <span class="TopBarLeft">
                {time()}
            </span>
            <span>
                {word()}
            </span>
        </div>
    );
}

export default TopBar;