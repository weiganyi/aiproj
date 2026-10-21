<template>
    <view class="navbar" :style="[navbarStyle]" :class="{ 'border-bottom': borderBottom }">
        <view class="status-bar" :style="{ height: statusBarHeight + 'px' }"></view>
        <view class="navbar-inner" :style="{ height: height + 'px' }">
            <view class="back-wrap" v-if="isBack" @tap="goBack">
                <uni-icons :type="backIconName" :color="backIconColor" :size="backIconSize"></uni-icons>
                <view class="back-text" v-if="backText" :style="[backTextStyle]">{{ backText }}</view>
            </view>
            <view class="navbar-content-title" :style="[titleStyle]">
                <view class="title" :style="{
						color: titleColor,
						fontSize: titleSize + 'rpx'
					}">
                    {{ title }}
                </view>
            </view>
            <view class="slot-right">
                <slot name="right"></slot>
            </view>
        </view>
    </view>
</template>

<script>
    // 获取系统状态栏的高度
    let systemInfo = uni.getSystemInfoSync();
    /**
     * navbar 自定义导航栏
     * @property {String Number} height 导航栏高度(不包括状态栏高度在内，内部自动加上)，注意这里的单位是px（默认44）
     * @property {String} back-icon-color 左边返回图标的颜色（默认#606266）
     * @property {String} back-icon-name 左边返回图标的名称，只能为uni-icon的图标（默认arrowleft）
     * @property {String Number} back-icon-size 左边返回图标的大小，单位rpx（默认24）
     * @property {String} back-text 返回图标右边的辅助提示文字
     * @property {Object} back-text-style 返回图标右边的辅助提示文字的样式，对象形式（默认{ color: '#606266' }）
     * @property {String} title 导航栏标题，如设置为空字符，将会隐藏标题占位区域
     * @property {String Number} title-width 导航栏标题的最大宽度，内容超出会以省略号隐藏，单位rpx（默认250）
     * @property {String} title-color 标题的颜色（默认#606266）
     * @property {String Number} title-size 导航栏标题字体大小，单位rpx（默认32）
     * @property {String Number} z-index 固定在顶部时的z-index值（默认980）
     * @property {Boolean} is-back 是否显示导航栏左边返回图标和辅助文字（默认true）
     * @property {Object} background 导航栏背景设置，见官网说明（默认{ background: '#ffffff' }）
     * @property {Boolean} border-bottom 导航栏底部是否显示下边框，如定义了较深的背景颜色，可取消此值（默认true）
     */
    export default {
        name: "navbar",
        props: {
            // 导航栏高度，单位px，非rpx
            height: {
                type: [String, Number],
                default: '44'
            },
            // 返回箭头的颜色
            backIconColor: {
                type: String,
                default: '#606266'
            },
            // 左边返回的图标
            backIconName: {
                type: String,
                default: 'arrowleft'
            },
            // 左边返回图标的大小，rpx
            backIconSize: {
                type: [String, Number],
                default: '24'
            },
            // 返回的文字提示
            backText: {
                type: String,
                default: ''
            },
            // 返回的文字的 样式
            backTextStyle: {
                type: Object,
                default () {
                    return {
                        color: '#606266'
                    }
                }
            },
            // 导航栏标题
            title: {
                type: String,
                default: ''
            },
            // 标题的宽度，如果需要自定义右侧内容，且右侧内容很多时，可能需要减少这个宽度，单位rpx
            titleWidth: {
                type: [String, Number],
                default: '250'
            },
            // 标题的颜色
            titleColor: {
                type: String,
                default: '#606266'
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
                default () {
                    return {
                        background: '#ffffff'
                    }
                }
            },
            // 是否显示导航栏的下边框
            borderBottom: {
                type: Boolean,
                default: true
            },
            zIndex: {
                type: [String, Number],
                default: ''
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
                // 合并用户传递的背景色对象
                Object.assign(style, this.background);
                return style;
            },
            // 导航中间的标题的样式
            titleStyle() {
                let style = {};
                style.left = (systemInfo.windowWidth - uni.upx2px(this.titleWidth)) / 2 + 'px';
                style.right = (systemInfo.windowWidth - uni.upx2px(this.titleWidth)) / 2 + 'px';
                style.width = uni.upx2px(this.titleWidth) + 'px';
                return style;
            }
        },
        created() {},
        methods: {
            goBack() {
                uni.navigateBack();
            }
        }
    };
</script>

<style scoped lang="scss">
    .navbar {
        width: 100%;
        position: fixed;
        left: 0;
        right: 0;
        top: 0;
        z-index: 991;
    }

    .status-bar {
        width: 100%;
    }

    .navbar-inner {
        display: flex;
        justify-content: space-between;
        position: relative;
        align-items: center;
    }

    .back-wrap {
        display: flex;
        align-items: center;
        flex: 1;
        flex-grow: 0;
        padding: 14rpx 14rpx 14rpx 24rpx;
    }

    .back-text {
        padding-left: 4rpx;
        font-size: 30rpx;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
    }

    .navbar-content-title {
        display: flex;
        align-items: center;
        justify-content: center;
        flex: 1;
        position: absolute;
        left: 0;
        right: 0;
        height: 60rpx;
        text-align: center;
        flex-shrink: 0;
    }

    .title {
        line-height: 60rpx;
        font-size: 32rpx;
        flex: 1;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
    }
</style>