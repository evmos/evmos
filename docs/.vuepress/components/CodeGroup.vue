<template>
  <ClientOnly>
    <div class="theme-code-group">
      <div class="theme-code-group__nav-wrapper">
        <div class="theme-code-group__nav">
          <ul class="theme-code-group__ul">
            <li
              v-for="(tab, i) in codeTabs"
              :key="tab.title"
              class="theme-code-group__li"
            >
              <button
                class="theme-code-group__nav-tab"
                :class="{
                  'theme-code-group__nav-tab-active': i === activeCodeTabIndex,
                }"
                @click="changeCodeTab(i)"
              >
                {{ tab.title }}
              </button>
            </li>
          </ul>
        </div>
      </div>
      <slot />
      <pre
        v-if="codeTabs.length < 1"
        class="pre-blank"
      >// Make sure to add code blocks to your code group</pre>
    </div>
  </ClientOnly>
</template>

<script>
export default {
  name: 'CodeGroup',
  data () {
    return {
      codeTabs: [],
      activeCodeTabIndex: -1
    }
  },
  watch: {
    activeCodeTabIndex (index) {
      this.activateCodeTab(index)
    }
  },
  mounted () {
    this.loadTabs()
  },
  methods: {
    changeCodeTab (index) {
      this.activeCodeTabIndex = index
    },
    loadTabs () {
      this.codeTabs = (this.$slots.default || []).filter(slot => Boolean(slot.componentOptions)).map((slot, index) => {
        if (slot.componentOptions.propsData.active === '') {
          this.activeCodeTabIndex = index
        }
        return {
          title: slot.componentOptions.propsData.title,
          elm: slot.elm
        }
      })
      if (this.activeCodeTabIndex === -1 && this.codeTabs.length > 0) {
        this.activeCodeTabIndex = 0
      }
      this.activateCodeTab(this.activeCodeTabIndex)
    },
    activateCodeTab (index) {
      this.codeTabs.forEach(tab => {
        if (tab.elm) {
          tab.elm.classList.remove('theme-code-block__active')
        }
      })
      if (this.codeTabs[index].elm) {
        this.codeTabs[index].elm.classList.add('theme-code-block__active')
      }
    }
  }
}
</script>

<style lang="stylus" scoped>
  .theme-code-group {
    border-radius: 0px 0px 20px 20px;
  }
  .theme-code-group__nav-wrapper {
    position: relative;
  }
  .theme-code-group__nav {
    inset: 0;
    background: #2e3148;
    margin-bottom: -50px;
    border-radius: 10px 10px 0px 0px;
    border-bottom: 1px solid lightgray;
    overflow-wrap: anywhere;
  }
  .theme-code-group__ul {
    margin: auto 0;
    padding-left: 0;
    display: inline-flex;
    list-style: none;
  }
  .theme-code-group__li {
    margin-top: 0px;
    margin-bottom: -1px;
    margin-inline: 10px;
  }
  .theme-code-group__nav-tab {
    border: 0;
    padding: 10px;
    cursor: pointer;
    background-color: transparent;
    font-size: 0.85em;
    line-height: 1.4;
    color: dimgray;
    font-weight: 600;
    height: 100%;
  }
  .theme-code-group__nav-tab-active {
    border-bottom: white 1px solid;
    color: white;
  }
</style>
