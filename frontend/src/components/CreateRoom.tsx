import "./CreateRoom.css";
import { For, createSignal } from "solid-js";
import { DOMAIN } from "../App";
import { useNavigate } from "@solidjs/router";

const MAX_PLAYER_SETTINGS = Array.from({ length: 11 }, (_, index) => 2 + index);
const ROUNDS_SETTINGS = Array.from({ length: 6 }, (_, index) => 1 + index);
const DRAW_TIME_SETTINGS = Array.from({ length: 12 }, (_, index) => 15 * (index + 1));

interface Props {
    onClose: () => void;
}

interface RoomSettings {
    playerLimit: number;
    totalRounds: number;
    timeLimitSecs: number;
    customWordBank: string[];
    isPublic: boolean;
}

async function createRoom(settings: RoomSettings): Promise<[string, boolean]> {
    try {
        const resp = await fetch(`http://${DOMAIN}/rooms/create`,
            { method: "POST", mode: "cors", body: JSON.stringify(settings) });
        const json = await resp.json();
        if (resp.ok) {
            return [json["Code"] as string, false];
        } else {
            return [json["ErrorDesc"] as string, true];
        }
    } catch(ex) {
        return ["Failed to create the room", true];
    }
}

const CreateRoom = ({ onClose }: Props) => {
    const [maxPlayersIndex, setMaxPlayersIndex] = createSignal(0);
    const [roundsIndex, setRoundsIndex] = createSignal(0);
    const [drawingTimeIndex, setDrawingTimeIndex] = createSignal(0);
    const [customWords, setCustomWords] = createSignal<string[]>([]);
    const [isPublic, setIsPublic] = createSignal<boolean>(true);

    const navigate = useNavigate();

    return (
        <div>
            <div class="Bg" onClick={onClose}/>
            <div class="CreateRoom">
                <div class="Panel SettingsPanel">
                    <h2 class="SettingsTitle">
                        Create Room
                    </h2>
                    <div class="X" onClick={onClose}/>
                    <SettingsSelect
                        label="Max Players"
                        index={maxPlayersIndex()}
                        setIndex={setMaxPlayersIndex}
                        options={MAX_PLAYER_SETTINGS}
                    />
                    <SettingsSelect
                        label="Total Rounds"
                        index={roundsIndex()}
                        setIndex={setRoundsIndex}
                        options={ROUNDS_SETTINGS}
                    />
                    <SettingsSelect
                        label="Drawing Time"
                        index={drawingTimeIndex()}
                        setIndex={setDrawingTimeIndex}
                        options={DRAW_TIME_SETTINGS}
                    />
                    <div class="SettingsLabel">
                        Custom Word Bank
                    </div>
                    <textarea
                        rows="3"
                        class="TextArea"
                        value={customWords().join(", ")}
                        onChange={(e) => {
                            const words = e.target.value
                                .replace(/\s/g, "")
                                .split(",");
                            setCustomWords(words);
                        }}
                    />
                    <div class="CheckboxContainer">
                        <label for="public">
                            Public Room
                        </label>
                        <input
                            class="Checkbox"
                            type="checkbox"
                            name="public"
                            checked={isPublic()}
                            onChange={() => setIsPublic(!isPublic())}
                        />
                    </div>
                    <button
                        class="Button SettingsButton"
                        onClick={async () => {
                            const settings: RoomSettings = {
                                playerLimit: MAX_PLAYER_SETTINGS[maxPlayersIndex()],
                                totalRounds: ROUNDS_SETTINGS[roundsIndex()],
                                timeLimitSecs: DRAW_TIME_SETTINGS[drawingTimeIndex()],
                                customWordBank: customWords(),
                                isPublic: isPublic()
                            };
                            const [value, error] = await createRoom(settings);
                            if (!error) {
                                navigate(`/rooms/${value}`);
                            } else {
                                console.log(value);
                            }
                        }}
                    >
                        Create Room!
                    </button>
                </div>
            </div>
        </div>
    );
}

interface SettingsSelectProps {
    label: string
    index: number;
    setIndex: (i: number) => void;
    options: number[];
}

const SettingsSelect = ({ label, index, setIndex, options }: SettingsSelectProps) => {
    return (
        <>
            <div class="SettingsLabel">
                { label }
            </div>
            <select
                class="Input SettingsInput"
                value={options[index]}
                onChange={e => setIndex(Number(e.target.value))}
            >
                <For each={options}>
                    {(i) => <option class="SelectOption" value={i}> {i} </option>}
                </For>
            </select>
        </>
    );
}

export default CreateRoom;