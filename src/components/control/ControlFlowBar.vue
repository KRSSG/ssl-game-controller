<template>
    <div class="control-flaw-bar">
        <div v-b-tooltip.hover
             :title="'Immediately stop all robots (' + Object.keys(keymapHalt)[0] + ')'">
            <b-button v-hotkey="keymapHalt"
                      ref="btnHalt"
                      v-on:click="send('halt')"
                      v-bind:disabled="halted">
                Halt
            </b-button>
        </div>

        <div v-b-tooltip.hover
              :title="'Continue based on last game event (' + Object.keys(keymapContinue)[0] + ')'">
            <b-button v-hotkey="keymapContinue"
                      ref="btnContinue"
                      v-on:click="triggerContinue"
                      v-bind:disabled="!continuePossible">
                Continue
                <span v-if="nextCommand !== ''">with</span>
                <span :class="{'team-blue': nextCommandFor === 'Blue', 'team-yellow': nextCommandFor === 'Yellow'}">
                    {{nextCommand}}
                </span>
            </b-button>
        </div>
    </div>
</template>

<script>
    import {isNonPausedStage, isPreStage} from "../../refereeState";

    export default {
        name: "ControlFlowBar",
        methods: {
            send: function (command) {
                this.$socket.sendObj({command: {commandType: command}})
            },
            triggerContinue() {
                this.$socket.sendObj({trigger: {triggerType: 'continue'}})
            },
        },
        computed: {
            keymapHalt() {
                return {
                    'esc': () => {
                        if (!this.$refs.btnHalt.disabled) {
                            this.send('halt')
                        }
                    }
                }
            },
            keymapContinue() {
                return {
                    'ctrl+space': () => {
                        if (!this.$refs.btnContinue.disabled) {
                            this.triggerContinue()
                        }
                    }
                }
            },
            state() {
                return this.$store.state.refBoxState
            },
            halted() {
                return this.state.command === 'halt';
            },
            continuePossible() {
                return this.nextCommand !== '';
            },
            stopAllowed() {
                return isNonPausedStage(this.$store.state.refBoxState)
                    || isPreStage(this.$store.state.refBoxState);
            },
            nextCommand() {
                if (this.halted) {
                    return 'stop'
                }
                return this.state.nextCommand;
            },
            nextCommandFor() {
                if (this.halted) {
                    return ''
                }
                return this.state.nextCommandFor;
            }
        }
    }
</script>

<style scoped>

    .control-flaw-bar {
        width: 100%;
        position: fixed;
        bottom: 0;
        text-align: center;
        display: flex;
        justify-content: center;
    }
</style>
