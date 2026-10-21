"use strict";
const common_vendor = require("../../common/vendor.js");
let systemInfo = common_vendor.index.getSystemInfoSync();
const _sfc_main = {
  name: "navbar",
  props: {
    // 导航栏高度，单位px，非rpx
    height: {
      type: [String, Number],
      default: "44"
    },
    // 返回箭头的颜色
    backIconColor: {
      type: String,
      default: "#606266"
    },
    // 左边返回的图标
    backIconName: {
      type: String,
      default: "arrowleft"
    },
    // 左边返回图标的大小，rpx
    backIconSize: {
      type: [String, Number],
      default: "24"
    },
    // 返回的文字提示
    backText: {
      type: String,
      default: ""
    },
    // 返回的文字的 样式
    backTextStyle: {
      type: Object,
      default() {
        return {
          color: "#606266"
        };
      }
    },
    // 导航栏标题
    title: {
      type: String,
      default: ""
    },
    // 标题的宽度，如果需要自定义右侧内容，且右侧内容很多时，可能需要减少这个宽度，单位rpx
    titleWidth: {
      type: [String, Number],
      default: "250"
    },
    // 标题的颜色
    titleColor: {
      type: String,
      default: "#606266"
    },
    // 标题的字体大小
    titleSize: {
      type: [String, Number],
      default: 32
    },
    isBack: {
      type: [Boolean, String],
      default: true
    },
    // 对象形式，因为用户可能定义一个纯色，或者线性渐变的颜色
    background: {
      type: Object,
      default() {
        return {
          background: "#ffffff"
        };
      }
    },
    // 是否显示导航栏的下边框
    borderBottom: {
      type: Boolean,
      default: true
    },
    zIndex: {
      type: [String, Number],
      default: ""
    }
  },
  data() {
    return {
      statusBarHeight: systemInfo.statusBarHeight
    };
  },
  computed: {
    // 整个导航栏的样式
    navbarStyle() {
      let style = {};
      style.zIndex = this.zIndex ? this.zIndex : 999;
      Object.assign(style, this.background);
      return style;
    },
    // 导航中间的标题的样式
    titleStyle() {
      let style = {};
      style.left = (systemInfo.windowWidth - common_vendor.index.upx2px(this.titleWidth)) / 2 + "px";
      style.right = (systemInfo.windowWidth - common_vendor.index.upx2px(this.titleWidth)) / 2 + "px";
      style.width = common_vendor.index.upx2px(this.titleWidth) + "px";
      return style;
    }
  },
  created() {
  },
  methods: {
    goBack() {
      common_vendor.index.navigateBack();
    }
  }
};
if (!Array) {
  const _easycom_uni_icons2 = common_vendor.resolveComponent("uni-icons");
  _easycom_uni_icons2();
}
const _easycom_uni_icons = () => "../uni-icons/uni-icons.js";
if (!Math) {
  _easycom_uni_icons();
}
function _sfc_render(_ctx, _cache, $props, $setup, $data, $options) {
  return common_vendor.e({
    a: $data.statusBarHeight + "px",
    b: $props.isBack
  }, $props.isBack ? common_vendor.e({
    c: common_vendor.p({
      type: $props.backIconName,
      color: $props.backIconColor,
      size: $props.backIconSize
    }),
    d: $props.backText
  }, $props.backText ? {
    e: common_vendor.t($props.backText),
    f: common_vendor.s($props.backTextStyle)
  } : {}, {
    g: common_vendor.o((...args) => $options.goBack && $options.goBack(...args))
  }) : {}, {
    h: common_vendor.t($props.title),
    i: $props.titleColor,
    j: $props.titleSize + "rpx",
    k: common_vendor.s($options.titleStyle),
    l: $props.height + "px",
    m: common_vendor.s($options.navbarStyle),
    n: $props.borderBottom ? 1 : ""
  });
}
const Component = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["render", _sfc_render], ["__scopeId", "data-v-eaf4c2e5"]]);
wx.createComponent(Component);
