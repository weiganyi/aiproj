<template>
    <view class="content">
        <navigator url="../index/index" open-type="switchTab">
            <button type="primary">回到首页</button>
        </navigator>
        <view>----------------------------------------</view>
        <avatar class="logo" :src="src" size="large" mode="circle"></avatar>
        <view class="text-area">
            <text class="title">{{title}}</text>
        </view>
        <view>----------------------------------------</view>
    </view>
</template>

<script>
    export default {
        data() {
            return {
                src: '/static/logo.png',
                title: '头像插件'
            }
        },
        onLoad() {

        },
        methods: {

        }
    }
</script>

<style>
    .content {
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
    }

    .logo {
        margin-top: 10rpx;
        margin-left: auto;
        margin-right: auto;
        margin-bottom: 50rpx;
    }

    .text-area {
        display: flex;
        justify-content: center;
    }

    .title {
        font-size: 36rpx;
        color: #8f8f94;
    }
</style>