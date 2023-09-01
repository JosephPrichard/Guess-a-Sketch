import "./Settings.css";
import { RoomProps } from "../../pages/room/Room";
import { START_CODE } from "../../pages/room/messages";

const Settings = ({ room }: RoomProps) => {
    room.subscribe(START_CODE, (payload) => {

    });

    return (
        <div>
            
        </div>
    );
}