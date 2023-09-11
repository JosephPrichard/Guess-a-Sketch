import "./Settings.css";
import { RoomProps } from "../../pages/room/Room";
import { OPTIONS_CODE, START_CODE } from "../../pages/room/messages";
import { For, createSignal } from "solid-js";

const MAX_PLAYER_SETTINGS = Array.from({ length: 11 }, (_, index) => 2 + index);
const ROUNDS_SETTINGS = Array.from({ length: 6 }, (_, index) => 1 + index);
const DRAW_TIME_SETTINGS = Array.from({ length: 12 }, (_, index) => 15 * (index + 1));

const Settings = ({ room }: RoomProps) => {
    const [maxPlayersIndex, setMaxPlayersIndex] = createSignal(0);
    const [roundsIndex, setRoundsIndex] = createSignal(0);
    const [drawingTimeIndex, setDrawingTimeIndex] = createSignal(0);
    const [customWords, setCustomWords] = createSignal<string[]>([]);

    room.subscribe(START_CODE, (payload) => {

    });

    return (
        <div class="SettingsWrapper">
            <div class="Panel SettingsPanel">
                <h2 class="SettingsTitle">
                    Room Options
                </h2>
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
                <button class="Button SettingsButton">
                    Start!
                </button>
            </div>
            <div class="Panel SettingsPanel">
                <h2 class="SettingsTitle">
                    Users Joined
                </h2>
                <div class="SettingsPlayers">
                    <For each={room.players()}>
                        {(player) => (
                            <div class="SettingsPlayer">
                                { player.name }
                            </div>
                        )}
                    </For>
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
                class="Input"
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

export default Settings;