<template>
    <div>
        <TeamSelection :model="model" label="By: "/>
        <b-button variant="primary"
                  @click="sendEvent()"
                  :disabled="model.team === null">
            Add
        </b-button>
    </div>
</template>

<script>
    import TeamSelection from "@/components/common/TeamSelection";

    export default {
        name: "PlacementFailed",
        components: {TeamSelection},
        data() {
            return {
                model: {
                    team: null,
                }
            }
        },
        methods: {
            sendEvent: function () {
                this.$socket.sendObj({
                    gameEvent: {
                        type: 'placementFailed',
                        details: {
                            ['placementFailed']: {
                                by_team: this.model.team.toLocaleUpperCase(),
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
