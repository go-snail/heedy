<template>
  <div>
    <v-row v-if="!$vuetify.breakpoint.xs">
      <v-col
        v-for="s in objects"
        :key="s.id"
        cols="12"
        xs="12"
        sm="6"
        md="4"
        lg="3"
        xl="3"
      >
        <v-hover #default="{ hover }">
          <v-card :elevation="hover ? 4 : 2">
            <v-list-item two-line subheader :to="`/objects/${s.id}`">
              <v-list-item-avatar>
                <h-icon
                  :image="s.icon"
                  :defaultIcon="defaultIcon"
                  :colorHash="s.id"
                ></h-icon>
              </v-list-item-avatar>
              <div
                v-if="showApps && s.app != null"
                style="position: absolute; bottom: 10px"
              >
                <h-icon
                  :image="app(s).icon"
                  defaultIcon="settings_input_component"
                  :colorHash="s.app"
                  :size="15"
                ></h-icon>
              </div>
              <v-list-item-content>
                <v-list-item-title>{{ s.name }}</v-list-item-title>
                <v-list-item-subtitle>{{ s.description }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-card>
        </v-hover>
      </v-col>
    </v-row>
    <!-- The above looks great on large screens, but nasty on mobile. We therefore create
    a custom mobile view-->
    <v-card v-else>
      <v-container fluid>
        <v-row no-gutters>
          <v-col cols="12" xs="12" v-for="s in objects" :key="s.id">
            <v-card class="pa-2" outlined tile>
              <v-list-item two-line subheader :to="`/objects/${s.id}`">
                <v-list-item-avatar>
                  <h-icon
                    :image="s.icon"
                    :defaultIcon="defaultIcon"
                    :colorHash="s.id"
                  ></h-icon>
                </v-list-item-avatar>
                <div
                  v-if="showApps && s.app != null"
                  style="position: absolute; bottom: 10px"
                >
                  <h-icon
                    :image="s.icon"
                    defaultIcon="settings_input_component"
                    :colorHash="s.app"
                    :size="15"
                  ></h-icon>
                </div>
                <v-list-item-content>
                  <v-list-item-title>{{ s.name }}</v-list-item-title>
                  <v-list-item-subtitle>{{
                    s.description
                  }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </div>
</template>
<script>
export default {
  props: {
    objects: Array,

    showApps: Boolean,
    defaultIcon: {
      type: String,
      default: "brightness_1",
    },
  },
  methods: {
    app(obj) {
      let empty_app = {
        id: obj.id,
        icon: "settings_input_component",
      };
      if (this.$store.state.heedy.apps == null) {
        return empty_app;
      }
      return this.$store.state.heedy.apps[obj.app] || empty_app;
    },
  },
  created() {
    if (this.showApps) {
      this.$store.dispatch("listApps");
    }
  },
};
</script>