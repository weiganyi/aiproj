"use strict";
const common_vendor = require("../../common/vendor.js");
const _sfc_main = {
  data() {
    return {
      src: "/static/logo.png",
      title: "头像插件"
    };
  },
  onLoad() {
  },
  methods: {}
};
if (!Array) {
  const _easycom_avatar2 = common_vendor.resolveComponent("avatar");
  _easycom_avatar2();
}
const _easycom_avatar = () => "../../components/avatar/avatar.js";
if (!Math) {
  _easycom_avatar();
}
function _sfc_render(_ctx, _cache, $props, $setup, $data, $options) {
  return {
    a: common_vendor.p({
      src: $data.src,
      size: "large",
      mode: "circle"
    }),
    b: common_vendor.t($data.title)
  };
}
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["render", _sfc_render]]);
wx.createPage(MiniProgramPage);
