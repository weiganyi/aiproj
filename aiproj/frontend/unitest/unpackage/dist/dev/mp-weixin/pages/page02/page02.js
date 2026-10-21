"use strict";
const common_vendor = require("../../common/vendor.js");
const _sfc_main = {
  data() {
    return {
      indexPicker: 0,
      arrayPicker: ["男", "女", "未知"],
      indexRatio: "BOY",
      itemsRatio: [
        {
          name: "BOY",
          value: "男",
          checked: "true"
        },
        {
          name: "GIRL",
          value: "女"
        },
        {
          name: "OTHER",
          value: "未知"
        }
      ]
    };
  },
  onLoad() {
  },
  methods: {
    bindPickerChange: function(e) {
      this.indexPicker = e.detail.value;
    },
    radioChange: function(e) {
      this.indexRatio = e.detail.value;
    }
  }
};
if (!Array) {
  const _easycom_navbar2 = common_vendor.resolveComponent("navbar");
  _easycom_navbar2();
}
const _easycom_navbar = () => "../../components/navbar/navbar.js";
if (!Math) {
  _easycom_navbar();
}
function _sfc_render(_ctx, _cache, $props, $setup, $data, $options) {
  return {
    a: common_vendor.p({
      ["back-text"]: "返回",
      title: "page02"
    }),
    b: common_vendor.t($data.arrayPicker[$data.indexPicker]),
    c: common_vendor.o((...args) => $options.bindPickerChange && $options.bindPickerChange(...args)),
    d: $data.indexPicker,
    e: $data.arrayPicker,
    f: common_vendor.f($data.itemsRatio, (item, indexRatio, i0) => {
      return {
        a: item.name,
        b: item.checked,
        c: common_vendor.t(item.value),
        d: item.name
      };
    }),
    g: common_vendor.o((...args) => $options.radioChange && $options.radioChange(...args)),
    h: common_vendor.t($data.indexRatio)
  };
}
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["render", _sfc_render]]);
wx.createPage(MiniProgramPage);
