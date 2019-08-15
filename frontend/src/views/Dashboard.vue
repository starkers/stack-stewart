<template>
  <div class="dashboard">
    <v-card >
      <v-card-title>

        <v-text-field
          v-model="search"
          append-icon="search"
          label="Search"
          single-line
          hide-details
        ></v-text-field>

        <v-spacer></v-spacer>

      </v-card-title>

      <v-data-table
        :headers="headers"
        :items="stacks"
        :search="search"
        class="elevation-3"
        :rows-per-page-items="[100, 200, 300, 400]"
        dark
      >

        <template slot="items" slot-scope="props">
          <td>{{ props.item.lane}}</td>
          <td>{{ props.item.name}}</td>

          <td>
            <!-- <v-tooltip up>
              <template v-slot:activator="{ on2 }"> -->

                <!-- <v-row align="center" v-for="c in props.item.containers" v-bind:key="c.name" v-on="on2"> -->
                <v-row align="center" v-for="c in props.item.containers" v-bind:key="c.name">
                  <v-col class="text-center" cols="12" sm="4" >
                    {{ c.tag }}
                  </v-col>
                  <v-col>
                    ({{ c.name }})
                    <br>
                  </v-col>
                </v-row>

              <!-- </template> -->

              <!-- <span v-for="c in props.item.containers" v-bind:key="c.name">
                {{c.image}}<br>
              </span>
            </v-tooltip> -->

          </td>

          <td>
            <v-row align="center">
              <v-col class="text-center" cols="12" sm="4">
                <div class="my-2">
                  <v-tooltip right>
                    <template v-slot:activator="{ on }">
                      <!-- TODO: change chip colours on errors -->
                      <v-chip
                        class="ma-2"
                        color="darken-1"
                        text-color="white"
                        v-on="on"
                      >
                        {{props.item.replicas.available}}-{{props.item.replicas.updated}}-{{props.item.replicas.ready}}
                        <v-icon right>check_circle</v-icon>
                      </v-chip>
                    </template>
                    <span>
                      Available: {{props.item.replicas.available}}<br>
                      Updated: {{props.item.replicas.updated}}<br>
                      Ready: {{props.item.replicas.ready}}<br>
                    </span>
                  </v-tooltip>

                </div>
              </v-col>
            </v-row>
          </td>
          <td>{{ props.item.agent}}</td>
          <td>{{ props.item.trace}}</td>
        </template>

      </v-data-table>
    </v-card>

  </div>
</template>

<script>

// See: https://github.com/iamshaunjp/vuetify-playlist/blob/lesson-16/todo-ninja/src/views/Dashboard.vue#L23
export default {


  data() {
    return {
      search: '',
      itemsPerPage: 100,
      stacks: [],
      headers: [
        // {
        //   text: 'item',
        //   align: 'left',
        //   sortable: false,
        //   value: 'name',
        // },
        { text: 'lane', value: 'lane' },
        { text: 'name', value: 'name' },
        { text: 'tag (container)', value: 'containers' },
        { text: 'replicas', value: 'replicas', sortable: false },
        { text: 'cluster', value: 'agent' },
        { text: 'trace (id)', value: 'trace' },
      ],
    }
  },
  mounted() {
    this.getStacks()
  },
  methods: {
    async getStacks() {
      try {
        const response = await fetch('/stacks')
        const data = await response.json()
        console.log("hello world")
        this.stacks = data
      } catch (error) {
        console.error(error)
      }
    }
  }
}
</script>

<style>

.state.complete{
  border-left: 4px solid #3cd1c2;
}
.state.updating{
  border-left: 4px solid #ffaa2c;
}
.state.errored{
  border-left: 4px solid #f83e70;
}
.v-chip.complete{
  background: #3cd1c2;
}
.v-chip.updating{
  background: #ffaa2c;
}
.v-chip.errored{
  background: #f83e70;
}

</style>
