<template>
    <div>
        <p>
            Minor unsporting behavior results in a yellow card.
        </p>
        <TeamSelection :model="model" label="By: "/>
        <b-form-group label="Reason: " label-cols-horizontal>
            <input type="text"
                   v-model="model.reason"/>
        </b-form-group>
        <b-button variant="primary"
                  @click="sendEvent()"
                  :disabled="model.team === null || model.reason === null">
            Add
        </b-button>
    </div>
</template>

<script>
    import TeamSelection from "@/components/common/TeamSelection";

    export default {
        name: "UnsportingBehaviorMinor",
        components: {TeamSelection},
        data() {
            return {
                model: {
                    team: null,
                    reason: null,
                }
            }
        },
        methods: {
            sendEvent: function () {
                this.$socket.sendObj({
                    gameEvent: {
                        type: 'unsportingBehaviorMinor',
                        details: {
                            ['unsportingBehaviorMinor']: {
                                by_team: this.model.team.toLocaleUpperCase(),
                                reason: this.model.reason
                            }
                        }
                    }
                });
                this.$root.$emit('bv::hide::modal', 'new-event-modal');
            }
        },
    }
</script>

<style scoped>

</style>
