package com.atoma.app.di

import com.atoma.app.data.network.NetworkConnectivityObserver
import com.atoma.app.data.network.NetworkConnectivityObserverImpl
import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
abstract class AppModule {

    @Binds
    @Singleton
    abstract fun bindNetworkConnectivityObserver(
        impl: NetworkConnectivityObserverImpl
    ): NetworkConnectivityObserver
}
