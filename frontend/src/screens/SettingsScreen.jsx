import { send } from "../api/index.js";

const NUMBER_SETTINGS = [
    { key: "startMoney", label: "Starting money", hint: "How much money each player starts with." },
    { key: "roundMoney", label: "Money for passing start", hint: "Earned every time a player completes a full lap." },
];

const TOGGLE_SETTINGS = [
    { key: "vacationCash", label: "Vacation jackpot", hint: "Taxes stack up. Landing on vacation collects all of it." },
    { key: "auction", label: "Property auctions", hint: "Skipped properties go to auction for everyone." },
    { key: "noRentInPrison", label: "No rent in prison", hint: "Players in prison can’t collect rent." },
    { key: "evenBuildings", label: "Even building", hint: "Houses must be built evenly across a set." },
];

export default function SettingsScreen({ settings, setSettings, next, back }) {
    const update = (key, value) => setSettings({ ...settings, [key]: value });

    function create() {
        send({ action: "CREATE_ROOM", settings });
    }

    return (
        <>
            <h2>Game Settings</h2>

            {NUMBER_SETTINGS.map(({ key, label, hint }) => (
                <div className="setting" key={key}>
                    <label>{label}</label>
                    <p>{hint}</p>
                    <input
                        type="number"
                        value={settings[key]}
                        onChange={e => update(key, Number(e.target.value))}
                    />
                </div>
            ))}

            {TOGGLE_SETTINGS.map(({ key, label, hint }) => (
                <div className="setting" key={key}>
                    <label>
                        <input
                            type="checkbox"
                            checked={settings[key]}
                            onChange={e => update(key, e.target.checked)}
                        />
                        {label}
                    </label>
                    <p>{hint}</p>
                </div>
            ))}

            <div className="button-row">
                <button onClick={back}>Back</button>
                <button onClick={next}>Start game</button>
                <button onClick={create}>Create Game</button>
            </div>
        </>
    );
}
