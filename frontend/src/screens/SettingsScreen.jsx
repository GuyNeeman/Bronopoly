import {sendMsg} from "../api/index.js";

export default function SettingsScreen({ settings, setSettings, next, back }) {

    function create(){
            const payload = JSON.stringify({
                action: "CREATE_GAME",
                settings: settings
            });
            sendMsg(payload);
    }
    return (
        <>
            <h2>Game Settings</h2>

            <div className="setting">
                <label>Starting money</label>
                <p>
                    How much money each player starts with.
                </p>
                <input
                    type="number"
                    value={settings.startMoney}
                    onChange={(e) =>
                        setSettings({
                            ...settings,
                            startMoney: Number(e.target.value)
                        })
                    }
                />
            </div>

            <div className="setting">
                <label>Money for passing start</label>
                <p>
                    Earned every time a player completes a full lap.
                </p>
                <input
                    type="number"
                    value={settings.roundMoney}
                    onChange={(e) =>
                        setSettings({
                            ...settings,
                            roundMoney: Number(e.target.value)
                        })
                    }
                />
            </div>

            <div className="setting">
                <label>
                    <input
                        type="checkbox"
                        checked={settings.vacationCash}
                        onChange={(e) =>
                            setSettings({
                                ...settings,
                                vacationCash: e.target.checked
                            })
                        }
                    />
                    Vacation jackpot
                </label>
                <p>
                    Taxes stack up. Landing on vacation collects all of it.
                </p>
            </div>

            <div className="setting">
                <label>
                    <input
                        type="checkbox"
                        checked={settings.auction}
                        onChange={(e) =>
                            setSettings({
                                ...settings,
                                auction: e.target.checked
                            })
                        }
                    />
                    Property auctions
                </label>
                <p>
                    Skipped properties go to auction for everyone.
                </p>
            </div>

            <div className="setting">
                <label>
                    <input
                        type="checkbox"
                        checked={settings.noRentInPrison}
                        onChange={(e) =>
                            setSettings({
                                ...settings,
                                noRentInPrison: e.target.checked
                            })
                        }
                    />
                    No rent in prison
                </label>
                <p>
                    Players in prison can’t collect rent.
                </p>
            </div>

            <div className="setting">
                <label>
                    <input
                        type="checkbox"
                        checked={settings.evenBuilding}
                        onChange={(e) =>
                            setSettings({
                                ...settings,
                                evenBuildings: e.target.checked
                            })
                        }
                    />
                    Even building
                </label>
                <p>
                    Houses must be built evenly across a set.
                </p>
            </div>

            <button onClick={back}>Back</button>
            <button onClick={next}>Start game</button>
            <button onClick={create}>Create Game</button>
        </>
    );
}