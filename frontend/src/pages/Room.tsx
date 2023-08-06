import Chat from "../components/Chat";
import "./Room.css"
import { useParams } from "@solidjs/router";

const Room = () => {
    const { code } = useParams();

    return (
        <div>
            <Chat/>
        </div>
    );
}

export default Room;