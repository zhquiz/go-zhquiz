<template>
  <section>
    <div class="HanziPage">
      <form class="field" @submit.prevent="q = q0">
        <div class="control">
          <input
            v-model="q0"
            class="input"
            type="search"
            name="q"
            placeholder="Type here to search."
            aria-label="search"
          />
        </div>
      </form>

      <div class="columns">
        <div class="column is-6 entry-display">
          <div
            class="hanzi-display clickable font-han"
            @contextmenu.prevent="(evt) => openContext(evt, current, 'hanzi')"
          >
            {{ current }}
          </div>

          <div class="buttons has-addons">
            <button
              class="button"
              :disabled="i < 1"
              @click="i--"
              @keypress="i--"
            >
              Previous
            </button>
            <button
              class="button"
              :disabled="i > entries.length - 2"
              @click="i++"
              @keypress="i++"
            >
              Next
            </button>
          </div>
        </div>

        <div class="column is-6">
          <b-collapse class="card" animation="slide" :open="!!sub.length">
            <div
              slot="trigger"
              slot-scope="props"
              class="card-header"
              role="button"
            >
              <h2 class="card-header-title">Subcompositions</h2>
              <a role="button" class="card-header-icon">
                <fontawesome :icon="props.open ? 'caret-down' : 'caret-up'" />
              </a>
            </div>

            <div class="card-content">
              <span
                v-for="h in sub"
                :key="h"
                class="font-han clickable"
                @contextmenu.prevent="(evt) => openContext(evt, h, 'hanzi')"
              >
                {{ h }}
              </span>
            </div>
          </b-collapse>

          <b-collapse class="card" animation="slide" :open="!!sup.length">
            <div
              slot="trigger"
              slot-scope="props"
              class="card-header"
              role="button"
            >
              <h2 class="card-header-title">Supercompositions</h2>
              <a role="button" class="card-header-icon">
                <fontawesome :icon="props.open ? 'caret-down' : 'caret-up'" />
              </a>
            </div>

            <div class="card-content">
              <span
                v-for="h in sup"
                :key="h"
                class="font-han clickable"
                @contextmenu.prevent="(evt) => openContext(evt, h, 'hanzi')"
              >
                {{ h }}
              </span>
            </div>
          </b-collapse>

          <b-collapse class="card" animation="slide" :open="!!variants.length">
            <div
              slot="trigger"
              slot-scope="props"
              class="card-header"
              role="button"
            >
              <h2 class="card-header-title">Variants</h2>
              <a role="button" class="card-header-icon">
                <fontawesome :icon="props.open ? 'caret-down' : 'caret-up'" />
              </a>
            </div>

            <div class="card-content">
              <span
                v-for="h in variants"
                :key="h"
                class="font-han clickable"
                @contextmenu.prevent="(evt) => openContext(evt, h, 'hanzi')"
              >
                {{ h }}
              </span>
            </div>
          </b-collapse>

          <b-collapse class="card" animation="slide" :open="!!vocabs.length">
            <div
              slot="trigger"
              slot-scope="props"
              class="card-header"
              role="button"
            >
              <h2 class="card-header-title">Vocabularies</h2>
              <a role="button" class="card-header-icon">
                <fontawesome :icon="props.open ? 'caret-down' : 'caret-up'" />
              </a>
            </div>

            <div class="card-content">
              <div v-for="(v, i) in vocabs" :key="i" class="long-item">
                <span
                  class="clickable"
                  @contextmenu.prevent="
                    (evt) => openContext(evt, v.simplified, 'vocab')
                  "
                >
                  {{ v.simplified }}
                </span>

                <span
                  v-if="v.traditional"
                  class="clickable"
                  @contextmenu.prevent="
                    (evt) => openContext(evt, v.traditional, 'vocab')
                  "
                >
                  {{ v.traditional }}
                </span>

                <span class="pinyin">[{{ v.pinyin }}]</span>

                <span>{{ v.english }}</span>
              </div>
            </div>
          </b-collapse>

          <b-collapse class="card" animation="slide" :open="!!sentences.length">
            <div
              slot="trigger"
              slot-scope="props"
              class="card-header"
              role="button"
            >
              <h2 class="card-header-title">Sentences</h2>
              <a role="button" class="card-header-icon">
                <fontawesome :icon="props.open ? 'caret-down' : 'caret-up'" />
              </a>
            </div>

            <div class="card-content">
              <div v-for="(s, i) in sentences" :key="i" class="long-item">
                <span
                  class="clickable"
                  @contextmenu.prevent="
                    (evt) => openContext(evt, s.chinese, 'sentence')
                  "
                >
                  {{ s.chinese }}
                </span>

                <span>{{ s.english }}</span>
              </div>
            </div>
          </b-collapse>
        </div>
      </div>
    </div>

    <ContextMenu
      ref="context"
      :entry="selected.entry"
      :type="selected.type"
      :additional="additionalContext"
      :pinyin="sentenceDef.pinyin"
      :english="sentenceDef.english"
    />
  </section>
</template>

<script lang="ts">
import XRegExp from 'xregexp'
import { Component, Ref, Vue, Watch } from 'vue-property-decorator'
import ContextMenu from '@/components/ContextMenu.vue'
import { api } from '@/assets/api'
import toPinyin from 'chinese-to-pinyin'

@Component<HanziPage>({
  components: {
    ContextMenu
  }
})
export default class HanziPage extends Vue {
  @Ref() context!: ContextMenu

  entries: string[] = []
  i = 0

  sub: string[] = []
  sup: string[] = []
  variants: string[] = []
  vocabs: Record<string, unknown>[] = []
  sentences: {
    chinese: string;
    pinyin: string;
    english: string;
  }[] = []

  selected: {
    entry: string;
    type: string;
  } = {
    entry: '',
    type: ''
  }

  q0 = ''

  get sentenceDef () {
    if (this.selected.type !== 'sentence') {
      return {}
    }

    const it = this.sentences.find((it) => it.chinese === this.selected.entry)
    if (!it) {
      return {}
    }

    return {
      pinyin: {
        [this.selected.entry]: it.pinyin
      },
      english: {
        [this.selected.entry]: it.english
      }
    }
  }

  get q () {
    const q = this.$route.query.q
    return (Array.isArray(q) ? q[0] : q) || ''
  }

  set q (q: string) {
    this.$router.push({ query: { q } })
  }

  get current () {
    return this.entries[this.i]
  }

  async created () {
    this.q0 = this.q

    if (this.additionalContext[0]) {
      await this.additionalContext[0].handler()
    }

    this.onQChange(this.q0)
  }

  get additionalContext () {
    if (!this.q) {
      return [
        {
          name: 'Reload',
          handler: async () => {
            const {
              data: { result }
            } = await api.get<{
              result: string;
            }>('/api/hanzi/random')

            this.q0 = result
          }
        }
      ]
    }

    return []
  }

  openContext (
    evt: MouseEvent,
    entry = this.selected.entry,
    type = this.selected.type
  ) {
    this.selected = { entry, type }
    this.context.open(evt)
  }

  @Watch('q')
  async onQChange (q: string) {
    if (XRegExp('\\p{Han}').test(q)) {
      const qs = q.split('').filter((h) => XRegExp('\\p{Han}').test(h))
      this.$set(
        this,
        'entries',
        qs.filter((h, i) => qs.indexOf(h) === i)
      )
    } else {
      const r = await api.get<{
        result: {
          entry: string;
        }[];
      }>('/api/hanzi/q', {
        params: {
          q
        }
      })
      this.entries = r.data.result.map(({ entry }) => entry)
    }

    this.i = 0
    this.load()
  }

  @Watch('current')
  load () {
    if (this.current) {
      this.loadHanzi()
      this.loadVocab()
      this.loadSentences()
    } else {
      this.sub = []
      this.sup = []
      this.variants = []
      this.vocabs = []
      this.sentences = []
    }
  }

  async loadHanzi () {
    const r = await api
      .get('/api/hanzi', {
        params: {
          entry: this.current
        }
      })
      .then((r) => r.data)

    this.sub = [...r.sub]
    this.sup = [...r.sup]
    this.variants = [...r.variants]
  }

  async loadVocab () {
    const {
      data: { result }
    } = await api.get('/api/vocab/q', {
      params: {
        q: this.current
      }
    })

    this.$set(this, 'vocabs', result)
  }

  async loadSentences () {
    const {
      data: { result }
    } = await api.get<{
      result: {
        chinese: string;
        english: string;
      }[];
    }>('/api/sentence/q', {
      params: {
        q: this.current,
        type: 'hanzi',
        generate: 10,
        select: 'chinese,english'
      }
    })

    this.$set(
      this,
      'sentences',
      result.map((r) => {
        const out = {
          chinese: r.chinese,
          pinyin: toPinyin(r.chinese, { keepRest: true, toneToNumber: true }),
          english: r.english.split('\x1f')[0]
        }

        return out
      })
    )
  }
}
</script>

<style scoped>
.entry-display {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.entry-display .clickable {
  min-height: 1.5em;
  display: block;
}

.card {
  margin-bottom: 1rem;
}

.card-content {
  max-height: 250px;
  overflow: scroll;
}

.card-content .font-han {
  font-size: 50px;
  display: inline-block;
}

.long-item > span + span {
  margin-left: 1rem;
}

.long-item > .pinyin {
  min-width: 8rem;
}
</style>
